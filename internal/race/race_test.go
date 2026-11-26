package race

import (
	"testing"
	"time"
)

type clock struct{ t time.Time }

func (c *clock) now() time.Time { return c.t }

func newHub() (*Hub, *clock) {
	c := &clock{time.Unix(1000, 0)}
	n := 0
	return NewHub(func(lang string) (string, string) { n++; return lang + " text " + string(rune('0'+n)), "" }, c.now), c
}

func ready(seats ...*Seat) {
	for _, s := range seats {
		s.SetReady(true)
	}
}

func TestReadyAndCountdown(t *testing.T) {
	h, c := newHub()
	a := h.Join("alice", "english")
	b := h.Join("alice", "english")
	k := h.Join("kim", "korean")
	sb := b.Snapshot()
	if sb.State != Waiting || sb.Players[1].Name != "alice#2" || !sb.Players[1].Me || sb.Players[0].Me {
		t.Fatalf("joined, not ready: wait; names unique, me marked: %+v", sb)
	}
	if k.Snapshot().Text == sb.Text {
		t.Fatal("other language gets its own room")
	}

	ready(a)
	if a.Snapshot().State != Waiting {
		t.Fatal("one ready player must not start the countdown")
	}
	ready(b)
	if s := a.Snapshot(); s.State != Counting || !s.Start.Equal(c.t.Add(Countdown)) || !s.Me().Ready {
		t.Fatalf("two ready players start the countdown: %+v", s)
	}
	b.SetReady(false)
	if a.Snapshot().State != Waiting {
		t.Fatal("un-readying below two stops the countdown")
	}
	ready(b)
	d := h.Join("dave", "english") // joins during the countdown, never readies
	a.Report(0.5, 80, false)       // before the start: ignored

	c.t = c.t.Add(Countdown)
	s := a.Snapshot()
	if s.State != Running || len(s.Players) != 2 || s.Players[0].Progress != 0 {
		t.Fatalf("racing with only the ready players, early report dropped: %+v", s)
	}
	sd := d.Snapshot()
	if sd.State != Waiting || sd.Text == s.Text || len(sd.Players) != 1 || sd.Players[0].Name != "dave" {
		t.Fatalf("unready player moves on to the next room: %+v", sd)
	}
	if late := h.Join("carol", "english").Snapshot(); late.Text != sd.Text || len(late.Players) != 2 {
		t.Fatalf("late joiner meets dave in the next room: %+v", late)
	}
}

func TestFinishOrderAndLeaving(t *testing.T) {
	h, c := newHub()
	a, b, d := h.Join("a", "en"), h.Join("b", "en"), h.Join("d", "en")
	ready(a, b, d)
	c.t = c.t.Add(Countdown + 3*time.Second)
	b.Report(1, 90, true)
	c.t = c.t.Add(2 * time.Second)
	a.Report(0.7, 60, false)
	d.Leave()
	b.Report(0.2, 1, false) // after finishing: ignored
	p := a.Snapshot().Players
	if p[1].Place != 1 || p[1].Time != 3*time.Second || p[1].Speed != 90 {
		t.Fatalf("b finished first at 3s: %+v", p[1])
	}
	if p[0].Place != 0 || p[0].Progress != 0.7 || !p[2].Left || len(p) != 3 {
		t.Fatalf("a still typing, d left but keeps its lane: %+v", p)
	}
	a.Report(1, 70, true)
	if p := a.Snapshot().Players[0]; p.Place != 2 || p.Time != 5*time.Second {
		t.Fatalf("a second at 5s: %+v", p)
	}
}

func TestRoomFull(t *testing.T) {
	h, _ := newHub()
	first := h.Join("p", "en")
	for range MaxPlayers - 1 {
		h.Join("p", "en")
	}
	if n := len(first.Snapshot().Players); n != MaxPlayers {
		t.Fatalf("room has %d players", n)
	}
	if s := h.Join("p", "en").Snapshot(); len(s.Players) != 1 {
		t.Fatalf("ninth player needs a new room: %+v", s)
	}
}

func TestClean(t *testing.T) {
	for in, want := range map[string]string{
		"":                      "player",
		"\x1b[31mevil\x1b[0m":   "31mevil0m",
		"홍길동":                   "홍길동",
		"averyveryverylongname": "averyveryver",
		"a b/c":                 "abc",
	} {
		if got := clean(in); got != want {
			t.Errorf("clean(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestBotsAndRematch(t *testing.T) {
	h, c := newHub() // texts are 14 runes: bot60 (300 cpm) ends at 2.8s, bot40 at 4.2s
	a := h.Join("alice", "english")
	b := h.Join("bob", "english")
	a.AddBot()
	a.AddBot()
	if s := a.Snapshot(); s.State != Waiting || len(s.Players) != 4 || s.Players[2].Name != "bot40" || s.Players[3].Name != "bot60" {
		t.Fatalf("bots join ready, but never start a race without a ready person: %+v", s)
	}
	ready(a, b)
	c.t = c.t.Add(Countdown + 3*time.Second)
	a.Report(1, 60, true) // at 3s: after bot60, before bot40
	c.t = c.t.Add(2 * time.Second)
	p := b.Snapshot().Players
	if p[3].Place != 1 || p[3].Time != 2800*time.Millisecond || p[0].Place != 2 || p[2].Place != 3 || p[1].Place != 0 {
		t.Fatalf("places in finish order, bots included: %+v", p)
	}

	if b.Rematch() {
		t.Fatal("only a finished player can rematch")
	}
	old := a.Snapshot().Text
	if !a.Rematch() {
		t.Fatal("finished: rematch")
	}
	s := a.Snapshot()
	if s.Text == old || s.State != Counting || len(s.Players) != 3 || !s.Me().Ready || s.Players[0].Name != "bot40" {
		t.Fatalf("rematch: new text, ready, with the bots, counting down: %+v", s)
	}
	if p := b.Snapshot().Players; p[0].Left || p[0].Place != 2 {
		t.Fatalf("a's finished lane keeps its place in the old race: %+v", p[0])
	}
	carol := h.Join("carol", "english")
	if carol.Snapshot().Text == s.Text {
		t.Fatal("the rematch room is only for the old room's players")
	}

	carol.AddBot()
	carol.Leave()
	if s := h.Join("erin", "english").Snapshot(); len(s.Players) != 1 {
		t.Fatalf("bots leave with the last person: %+v", s)
	}
}
