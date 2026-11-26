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
// column's gradient color and returns the color to use, or false to hide it.
func (c canvas) art(x, y int, shade func(col int, fg tcell.Color) (tcell.Color, bool)) {
	from, to := titleColors(c.th)
	span := float64(logoWidth() + 2*(len(logo)-1)) // rows count double: cells are ~2:1
	for j, l := range logo {
		xx := x
		for _, r := range l {
			if r != ' ' {
				fg, show := blend(from, to, float64(xx-x+2*j)/span), true
				if shade != nil {
					fg, show = shade(xx-x, fg)
				}
				if show {
					c.put(xx, y+j, r, c.st(fg))
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
	keyFall    = 280 * time.Millisecond // a typed key drops into place from above the screen
	keyFlash   = 220 * time.Millisecond // a landed key lights up in the text color, then settles
)

// keyAt is when splash key k is typed: after a short pause, at an uneven,
// human-ish rhythm of 130-220ms per key.
func keyAt(k int) time.Duration {
	d := 200 * time.Millisecond
	for i := 1; i <= k; i++ {
		d += time.Duration(130+(i*61)%90) * time.Millisecond
	}
	return d
}

// splash types the name in as if on a keyboard: the logo is split into one
// slice per letter, each dropping in on its keystroke and flashing as it lands, behind a
// full-height caret that blinks once all are typed, the way a test looks
// before the first key. Without room for the logo, the name itself is typed.
func (c canvas) splash(w, h int, since time.Duration) {
	keys := len(splashName)
	typed := 0
	for typed < keys && since >= keyAt(typed) {
		typed++
	}
	caret := c.st(c.th.bg).Background(c.th.caret)
	showCaret := typed < keys || since/(300*time.Millisecond)%2 == 0
	if lw, lh := logoWidth(), len(logo); lw+2 <= w && lh <= h {
		x, y := min((w-lw)/2, w-lw-2), (h-lh)/2 // same column as the title, caret still fits
		for k := range typed {
			age := since - keyAt(k)
			p := min(float64(age)/float64(keyFall), 1)
			dy := int(math.Round(float64(-(y + lh)) * (1 - p*p))) // from above the screen, speeding up like a drop
			c.art(x, y+dy, func(col int, fg tcell.Color) (tcell.Color, bool) {
				if col*keys/lw != k {
					return fg, false
				}
				if land := age - keyFall; land >= 0 && land < keyFlash {
					fg = blend(c.th.text, fg, float64(land)/float64(keyFlash)) // flash on landing
				}
				return fg, true
			})
		}
		if showCaret {
			cx := x + typed*lw/keys
			if typed == keys {
				cx++ // rest one cell clear of the art
			}
			for j := range lh {
				c.put(cx, y+j, ' ', caret)
			}
		}
		return
	}
	x := c.print((w-keys-1)/2, h/2, splashName[:typed], c.st(c.th.text).Bold(true))
	if showCaret {
		c.put(x, h/2, ' ', caret)
	}
}
