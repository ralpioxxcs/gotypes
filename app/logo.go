package app

import (
	"regexp"
	"runtime/debug"
	"strings"

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
█▀▀▀▀ ▀ ▀▀▄ ▀▀█▀▀ ▀   █ █▀▀▀█ █▀▀▀ █▀▀▀▀
█ ▀▀▄ █   █   █   ▀▀▀▀▄ █▀▀▀▀ █▀   ▀▀▀▀▄
█▄▄▄▄ █▄▄▄█   ▄   ▄▄▄▄▄ ▄     █▄▄▄ ▄▄▄▄▄
`, "\n"), "\n")

// logo draws the title art centered above content starting at row y0, fading
// from main to sub left to right, with the version small at its lower right;
// skipped when the screen has no room for it.
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
	for j, l := range logo {
		xx := x
		for _, r := range l {
			if r != ' ' {
				c.put(xx, y+j, r, c.st(blend(c.th.main, c.th.sub, 0.6*float64(xx-x)/float64(lw))))
			}
			xx += rw(r)
		}
	}
	if x+lw+2+len(buildVersion) <= w {
		c.print(x+lw+2, y+len(logo)-1, buildVersion, c.st(c.th.sub))
	}
}
