package app

import (
	"math"
	"regexp"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

var buildVersion = version()

func version() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}
	var rev string
	var dirty bool
	for _, s := range bi.Settings {
		switch s.Key {
		case "vcs.revision":
			rev = s.Value
		case "vcs.modified":
			dirty = s.Value == "true"
		}
	}
	return shortVersion(bi.Main.Version, rev, dirty)
}

var pseudo = regexp.MustCompile(`-(0\.)?\d{14}-[0-9a-f]{12}`) // v0.0.0-20060102150405-abcdef123456

// shortVersion is the tag when built from one, else the short commit (with *
// for uncommitted changes), else "dev".
func shortVersion(v, rev string, dirty bool) string {
	switch {
	case v != "" && v != "(devel)" && !pseudo.MatchString(v):
		return strings.TrimSuffix(v, "+dirty")
	case len(rev) >= 7 && dirty:
		return rev[:7] + "*"
	case len(rev) >= 7:
		return rev[:7]
	}
	return "dev"
}

var logo = strings.Split(strings.Trim(`
┌───┐┌───┐┌───┐┌───┐┌───┐┌───┐┌───┐
│ g ││ o ││ t ││ y ││ p ││ e ││ s │
└───┘└───┘└───┘└───┘└───┘└───┘└───┘
`, "\n"), "\n")

// logo draws the title art centered above content starting at row y0, with a
// diagonal gradient from main to main's hue turned 90°, and the version small
// at its lower right; skipped when the screen has no room for it.
func (c canvas) logo(w, y0 int) {
	y, lw := y0-len(logo)-1, logoWidth()
	if y < 1 || lw > w {
		return
	}
	x := (w - lw) / 2
	c.art(x, y, nil)
	if x+lw+2+len(buildVersion) <= w {
		c.print(x+lw+2, y+len(logo)-1, buildVersion, c.st(c.th.sub))
	}
}

func logoWidth() int {
	lw := 0
	for _, l := range logo {
		lw = max(lw, runewidth.StringWidth(l))
	}
	return lw
}

// chroma is how colorful c is, 0 (grey) to 1.
func chroma(c tcell.Color) float64 {
	r, g, b := c.RGB()
	return float64(max(r, g, b)-min(r, g, b)) / 255
}

// titleColors are the gradient ends: the theme's first colorful role (main
// unless it is grey, as in dracula) to that hue turned 90°; all-grey themes
// fade main to sub.
func titleColors(th theme) (from, to tcell.Color) {
	for _, c := range []tcell.Color{th.main, th.caret, th.sub, th.err} {
		if chroma(c) >= 0.25 {
			return c, hueShift(c, 90)
		}
	}
	return th.main, th.sub
}

// art draws the logo at x, y in the title gradient. shade, if set, gets each
// cell's column, rune and gradient color and returns the rune and color to
// use, or false to hide it.
func (c canvas) art(x, y int, shade func(col int, r rune, fg tcell.Color) (rune, tcell.Color, bool)) {
	from, to := titleColors(c.th)
	span := float64(logoWidth() + 2*(len(logo)-1)) // rows count double: cells are ~2:1
	for j, l := range logo {
		xx := x
		for _, r := range l {
			if r != ' ' {
				fg, show := blend(from, to, float64(xx-x+2*j)/span), true
				out := r
				if shade != nil {
					out, fg, show = shade(xx-x, r, fg)
				}
				if show {
					c.put(xx, y+j, out, c.st(fg))
				}
			}
			xx += rw(r)
		}
	}
}

// hueShift turns c's hue by deg degrees, keeping saturation and lightness.
func hueShift(c tcell.Color, deg float64) tcell.Color {
	r, g, b := c.RGB()
	rf, gf, bf := float64(r)/255, float64(g)/255, float64(b)/255
	mx, mn := max(rf, gf, bf), min(rf, gf, bf)
	l, d := (mx+mn)/2, mx-mn
	if d < 0.05 { // grey: hue is meaningless
		return c
	}
	s := d / (1 - math.Abs(2*l-1))
	var h float64
	switch mx {
	case rf:
		h = math.Mod((gf-bf)/d, 6)
	case gf:
		h = (bf-rf)/d + 2
	default:
		h = (rf-gf)/d + 4
	}
	h = math.Mod(h*60+deg+360, 360)
	ch := (1 - math.Abs(2*l-1)) * s
	xc := ch * (1 - math.Abs(math.Mod(h/60, 2)-1))
	var r1, g1, b1 float64
	switch {
	case h < 60:
		r1, g1 = ch, xc
	case h < 120:
		r1, g1 = xc, ch
	case h < 180:
		g1, b1 = ch, xc
	case h < 240:
		g1, b1 = xc, ch
	case h < 300:
		r1, b1 = xc, ch
	default:
		r1, b1 = ch, xc
	}
	m := l - ch/2
	to := func(v float64) int32 { return int32(math.Round((v + m) * 255)) }
	return tcell.NewRGBColor(to(r1), to(g1), to(b1))
}

const (
	splashName = "gotypes"
	splashTime = 2400 * time.Millisecond
	keyFall    = 280 * time.Millisecond // drop: a typed key falls into place from above the screen
	keyFlash   = 220 * time.Millisecond // a key lights up in the text color as it lands, then settles
	scramble   = 350 * time.Millisecond // scramble: a key cycles random letters this long
	waveTime   = 700 * time.Millisecond // wave: one shine sweep across the logo
	flyTime    = 400 * time.Millisecond // at the end the logo glides up to the title row
)

// splashStyles are the splash effects, one picked at random per start.
var splashStyles = []string{"drop", "typo", "scramble", "wave", "press"}

// splashSteps is what has been typed after each keystroke; the typo splash
// hits p too early, shows it wrong and backspaces over it.
func splashSteps(style string) []string {
	if style == "typo" {
		return []string{"g", "go", "got", "gotp", "got", "goty", "gotyp", "gotype", "gotypes"}
	}
	var s []string
	for i := range splashName {
		s = append(s, splashName[:i+1])
	}
	return s
}

// keyAt is when splash keystroke k lands: after a short pause, at an uneven,
// human-ish rhythm of 130-220ms per key.
func keyAt(k int) time.Duration {
	d := 200 * time.Millisecond
	for i := 1; i <= k; i++ {
		d += time.Duration(130+(i*61)%90) * time.Millisecond
	}
	return d
}

// splash types the name in as if on a keyboard: the logo is split into one
// slice per letter, each appearing on its keystroke with a flash, behind a
// full-height caret that blinks once all are typed, the way a test looks
// before the first key. style adds an effect; at the end the logo glides up to
// titleY, where the main screen draws it. Without room for the logo, the name
// itself is typed.
func (c canvas) splash(w, h int, since time.Duration, titleY int, style string) {
	steps := splashSteps(style)
	n := 0
	for n < len(steps) && since >= keyAt(n) {
		n++
	}
	buf := ""
	if n > 0 {
		buf = steps[n-1]
	}
	done, end, keys := n == len(steps), keyAt(len(steps)-1), len(splashName)
	caret := c.st(c.th.bg).Background(c.th.caret)
	showCaret := !done || since/(300*time.Millisecond)%2 == 0
	if lw, lh := logoWidth(), len(logo); lw+2 <= w && lh <= h {
		x, y := min((w-lw)/2, w-lw-2), (h-lh)/2 // same column as the title, caret still fits
		if t := since - (splashTime - flyTime); titleY >= 1 && t > 0 {
			p := float64(t) / float64(flyTime)
			y += int(math.Round(float64(titleY-y) * (1 - (1-p)*(1-p)))) // ease out
			showCaret = false
		}
		for k := range len(buf) {
			set := n - 1 // the keystroke that put buf[k] there
			for set > 0 && len(steps[set-1]) > k && steps[set-1][k] == buf[k] {
				set--
			}
			age := since - keyAt(set)
			letter, wrong := rune(buf[k]), buf[k] != splashName[k]
			dy, flash := 0, age
			switch style {
			case "", "drop":
				p := min(float64(age)/float64(keyFall), 1)
				dy, flash = int(math.Round(float64(-(y+lh))*(1-p*p))), age-keyFall // speeding up like a drop
			case "scramble":
				if age < scramble {
					letter = rune('a' + (k*7+int(since/(60*time.Millisecond))*11)%26)
				}
				flash = age - scramble
			case "press": // once typed, the keys go down one after another
				if t := since - end - 150*time.Millisecond - time.Duration(k)*80*time.Millisecond; t >= 0 && t < 120*time.Millisecond {
					dy = 1
				}
			}
			c.art(x, y+dy, func(col int, r rune, fg tcell.Color) (rune, tcell.Color, bool) {
				if col*keys/lw != k {
					return r, fg, false
				}
				if r == rune(splashName[k]) {
					r = letter
				}
				switch {
				case wrong:
					fg = c.th.err
				case style == "wave" && done: // a bright band sweeping left to right
					pos := float64((since-end)%waveTime)/float64(waveTime)*float64(lw+8) - 4
					fg = blend(c.th.text, fg, min(math.Abs(float64(col)-pos)/4, 1))
				}
				if flash >= 0 && flash < keyFlash {
					fg = blend(c.th.text, fg, float64(flash)/float64(keyFlash))
				}
				return r, fg, true
			})
		}
		if showCaret {
			cx := x + len(buf)*lw/keys
			if done {
				cx++ // rest one cell clear of the art
			}
			for j := range lh {
				c.put(cx, y+j, ' ', caret)
			}
		}
		return
	}
	x := c.print((w-keys-1)/2, h/2, buf, c.st(c.th.text).Bold(true))
	if showCaret {
		c.put(x, h/2, ' ', caret)
	}
}
