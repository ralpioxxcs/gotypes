package app

import (
	"slices"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/ralpioxxcs/gotypes/internal/race"
)

func TestRace(t *testing.T) {
	clk := time.Unix(0, 0)
	hub := race.NewHub(func(string) (string, string) { return "ab cd", "src" }, func() time.Time { return clk })
	join := func(name string, hub *race.Hub) *ui {
		u := &ui{lists: map[string][]string{"english": {"x"}}, langs: []string{"english"}, lang: "english",
			hub: hub, name: name, mode: modeIndex("race"), best: map[string]float64{}}
		u.restart()
		return u
	}
	typ := func(u *ui, s string) {
		for _, r := range s {
			u.key(tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone))
		}
		u.report()
	}

	if local := join("solo", nil); modes[local.mode].race || slices.Contains(categories[0].options(local), "race") {
		t.Fatal("no hub: race must not be offered")
	}

	a, b := join("a", hub), join("b", hub)
	if a.quote.Text != "ab cd" || len(b.e.Words) != 2 {
		t.Fatalf("both race the room's text: %q, %d words", a.quote.Text, len(b.e.Words))
	}
	enter := func(u *ui) { u.key(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone)) }
	c := join("c", hub) // never readies
	enter(a)
	enter(b)
	if s := a.seat.Snapshot(); s.State != race.Counting || !s.Me().Ready {
		t.Fatalf("enter readies; two ready players count down: %+v", s)
	}
	typ(a, "ab")
	if a.e.Started {
		t.Fatal("typing before the countdown ends must be ignored")
	}

	clk = clk.Add(race.Countdown + 4*time.Second)
	c.quote.Text = "stale" // every room gets "ab cd", so mark c's old text to see the switch
	c.report()
	if s := c.seat.Snapshot(); s.State != race.Waiting || len(s.Players) != 1 || c.quote.Text != "ab cd" {
		t.Fatalf("c was not ready: moved to a fresh room: %+v", s)
	}
	typ(a, "ab cd")
	typ(b, "ab ")
	p := b.seat.Snapshot().Players
	if !a.e.Finished || p[0].Place != 1 || p[0].Time != 4*time.Second || p[1].Progress != 0.5 {
		t.Fatalf("a should win at 4s while b is halfway: %+v", p)
	}

	b.key(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)) // b leaves for a new room
	if p := a.seat.Snapshot().Players; !p[1].Left {
		t.Fatalf("b's lane should show it left: %+v", p)
	}
	if s := b.seat.Snapshot(); s.State != race.Waiting || len(s.Players) != 2 || s.Players[0].Name != "c" {
		t.Fatalf("tab during a race joins the next room, where c waits: %+v", s)
	}
}
