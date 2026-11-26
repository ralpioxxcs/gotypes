package app

import (
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
			if len(cells[i].Runes) > 0 && cells[i].Runes[0] == '■' {
				return true
			}
		}
		return false
	}
	for _, tc := range []struct {
		w, y0 int
		want  bool
	}{{100, 10, true}, {100, 7, false}, {60, 10, false}} {
		scr.SetSize(tc.w, 30)
		scr.Clear()
		c.logo(tc.w, tc.y0)
		scr.Show()
		if has() != tc.want {
			t.Errorf("w=%d y0=%d: logo drawn = %v", tc.w, tc.y0, !tc.want)
		}
	}
}
