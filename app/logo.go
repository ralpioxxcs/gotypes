package app

import (
	"math"
	"regexp"
	"runtime/debug"
	"strings"

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
	y := y0 - len(logo) - 1
	lw := 0
	for _, l := range logo {
		lw = max(lw, runewidth.StringWidth(l))
	}
	if y < 1 || lw > w {
		return
	}
	x := (w - lw) / 2
	from, to := c.th.main, hueShift(c.th.main, 90)
	if to == from { // grey main: no hue to turn, fade to sub instead
		to = c.th.sub
	}
	span := float64(lw + 2*(len(logo)-1)) // rows count double: cells are ~2:1
	for j, l := range logo {
		xx := x
		for _, r := range l {
			if r != ' ' {
				c.put(xx, y+j, r, c.st(blend(from, to, float64(xx-x+2*j)/span)))
			}
			xx += rw(r)
		}
	}
	if x+lw+2+len(buildVersion) <= w {
		c.print(x+lw+2, y+len(logo)-1, buildVersion, c.st(c.th.sub))
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
