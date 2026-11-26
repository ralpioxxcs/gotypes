package app

import (
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/ralpioxxcs/gotypes/internal/engine"
)

func TestChartDrawsEndToEnd(t *testing.T) {
	scr := tcell.NewSimulationScreen("UTF-8")
	if err := scr.Init(); err != nil {
		t.Fatal(err)
	}
	scr.SetSize(100, 30)
	c := canvas{scr, themes[0]}
	for _, n := range []int{1, 2, 7, 60, 300} {
		ps := make([]engine.Second, n)
		for i := range ps {
			ps[i] = engine.Second{WPM: float64(40 + i%7*10), Raw: float64(60 + i%3*30)}
		}
		scr.Clear()
		g := c.chart(ps, -1, time.Hour, 0, 2, 80, 8)
		// hovering a column selects a second drawn in that column
		last := g.dw/2 - 1
		if g.n != n || g.second(g.x) != 0 || (n > 1 && g.col(g.second(g.x+last)) != last) {
			t.Fatalf("n=%d: geom %+v maps edges to %d..%d", n, g, g.second(g.x), g.second(g.x+last))
		}
		// the line must reach both the first and the last plot column
		for _, col := range []int{g.x, g.x + g.dw/2 - 1} {
			found := false
			for row := g.y; row < g.y+g.h; row++ {
				if r, _, _, _ := scr.GetContent(col, row); r >= 0x2801 && r <= 0x28ff {
					found = true
				}
			}
			if !found {
				t.Errorf("n=%d: no braille dots in column %d", n, col)
			}
		}
	}
}

func TestLogoOnlyWhenItFits(t *testing.T) {
	scr := tcell.NewSimulationScreen("UTF-8")
	if err := scr.Init(); err != nil {
		t.Fatal(err)
	}
	c := canvas{scr, themes[0]}
	has := func() bool {
		cells, w, h := scr.GetContents()
		for i := range w * h {
			if len(cells[i].Runes) > 0 && cells[i].Runes[0] == []rune(strings.TrimSpace(logo[0]))[0] {
				return true
			}
		}
		return false
	}
	for _, tc := range []struct {
		w, y0 int
		want  bool
	}{{100, 10, true}, {100, len(logo), false}, {20, 10, false}} {
		scr.SetSize(tc.w, 30)
		scr.Clear()
		c.logo(tc.w, tc.y0)
		scr.Show()
		if has() != tc.want {
			t.Errorf("w=%d y0=%d: logo drawn = %v", tc.w, tc.y0, !tc.want)
		}
	}
}

func TestShortVersion(t *testing.T) {
	for _, tc := range []struct {
		v, rev string
		dirty  bool
		want   string
	}{
		{"v1.2.0", "abcdef1234", false, "v1.2.0"},
		{"v1.2.0+dirty", "abcdef1234", true, "v1.2.0"},
		{"v0.0.0-20261126013012-a627968e7f1c", "a627968e7f1c00", false, "a627968"},
		{"v1.2.1-0.20261126013012-a627968e7f1c+dirty", "a627968e7f1c00", true, "a627968*"},
		{"(devel)", "", false, "dev"},
	} {
		if got := shortVersion(tc.v, tc.rev, tc.dirty); got != tc.want {
			t.Errorf("shortVersion(%q, %q, %v) = %q, want %q", tc.v, tc.rev, tc.dirty, got, tc.want)
		}
	}
}

func TestHueShift(t *testing.T) {
	for _, tc := range []struct{ in, want tcell.Color }{
		{tcell.NewRGBColor(255, 0, 0), tcell.NewRGBColor(0, 255, 0)},       // red +120° = green
		{tcell.NewRGBColor(51, 102, 204), tcell.NewRGBColor(204, 51, 102)}, // keeps s and l
		{tcell.NewRGBColor(238, 238, 238), tcell.NewRGBColor(238, 238, 238)},
	} {
		if got := hueShift(tc.in, 120); got != tc.want {
			r, g, b := got.RGB()
			t.Errorf("hueShift(%v) = %d,%d,%d", tc.in, r, g, b)
		}
	}
}
