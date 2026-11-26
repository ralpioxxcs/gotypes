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

func TestMatchAndCountdown(t *testing.T) {
	h, c := newHub()
	a := h.Join("alice", "english")
	if s := a.Snapshot(); s.State != Waiting || len(s.Players) != 1 || !s.Players[0].Me {
		t.Fatalf("alone should wait: %+v", s)
	}
	b := h.Join("alice", "english")
	k := h.Join("kim", "korean")
	sa, sb := a.Snapshot(), b.Snapshot()
	if sa.State != Counting || !sa.Start.Equal(c.t.Add(Countdown)) || sa.Text != sb.Text {
		t.Fatalf("second player starts the countdown on one text: %+v %+v", sa, sb)
	}
	if sb.Players[1].Name != "alice#2" || !sb.Players[1].Me || sb.Players[0].Me {
		t.Fatalf("names must be unique and me marked: %+v", sb.Players)
	}
	if sk := k.Snapshot(); sk.State != Waiting || sk.Text == sa.Text {
		t.Fatalf("other language gets its own room: %+v", sk)
	}

	b.Leave()
	if s := a.Snapshot(); s.State != Waiting || len(s.Players) != 1 {
		t.Fatalf("leaving before the start stops the countdown: %+v", s)
	}
	b = h.Join("bob", "english")
	a.Report(0.5, 80, false) // before the start: ignored
	c.t = c.t.Add(Countdown)
	if s := a.Snapshot(); s.State != Running || s.Players[0].Progress != 0 {
		t.Fatalf("running, early report dropped: %+v", s)
	}
	late := h.Join("carol", "english")
	if late.Snapshot().Text == a.Snapshot().Text {
		t.Fatal("joining a running race must open a new room")
	}
}

func TestFinishOrderAndLeaving(t *testing.T) {
	h, c := newHub()
	a, b, d := h.Join("a", "en"), h.Join("b", "en"), h.Join("d", "en")
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
