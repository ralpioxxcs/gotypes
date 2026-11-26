package app

import (
	"context"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
)

// waitFor polls cond for up to a second.
func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	for end := time.Now().Add(time.Second); time.Now().Before(end); time.Sleep(10 * time.Millisecond) {
		if cond() {
			return
		}
	}
	t.Fatalf("timed out waiting for %s", what)
}

// simScreen reports Play's first frame. The simulation screen's buffers are
// not safe to read while Play draws, so tests wait on this instead.
type simScreen struct {
	tcell.SimulationScreen
	shown chan struct{}
	once  *sync.Once
}

func newSim() simScreen {
	return simScreen{tcell.NewSimulationScreen("UTF-8"), make(chan struct{}), &sync.Once{}}
}

func (s simScreen) Show() {
	s.SimulationScreen.Show()
	s.once.Do(func() { close(s.shown) })
}

func started(t *testing.T, s simScreen) {
	t.Helper()
	select {
	case <-s.shown:
	case <-time.After(time.Second):
		t.Fatal("no frame drawn")
	}
}

func TestPlaySessionsEndWithoutLeaks(t *testing.T) {
	before := runtime.NumGoroutine()
	for range 20 {
		s := newSim()
		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan error)
		go func() { done <- Play(ctx, s, "", nil, "") }()
		started(t, s)
		cancel() // like an SSH client disconnecting
		if err := <-done; err != nil {
			t.Fatal(err)
		}
	}
	waitFor(t, "goroutines to settle", func() bool { return runtime.NumGoroutine() <= before+2 })
}

func TestPlayQuitsOnCtrlC(t *testing.T) {
	s := newSim()
	done := make(chan error)
	go func() { done <- Play(context.Background(), s, "", nil, "") }()
	started(t, s)
	s.InjectKey(tcell.KeyCtrlC, 0, tcell.ModNone)
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("ctrl+c did not end the session")
	}
}

func TestSplashSweepsLogoThenAnyKeySkips(t *testing.T) {
	scr := tcell.NewSimulationScreen("UTF-8")
	if err := scr.Init(); err != nil {
		t.Fatal(err)
	}
	u := &ui{lists: map[string][]string{"english": {"word"}}, langs: []string{"english"}, lang: "english", mode: 1}
	u.restart()
	screen := func(w, h int) string {
		scr.SetSize(w, h)
		u.draw(scr, w, h)
		scr.Show()
		cells, cw, ch := scr.GetContents()
		var b strings.Builder
		for i := range cw * ch {
			if len(cells[i].Runes) > 0 {
				b.WriteRune(cells[i].Runes[0])
			}
		}
		return b.String()
	}
	top := strings.TrimSpace(logo[0]) // whatever the art is
	first := string([]rune(top)[:1])
	for _, tc := range []struct {
		at        time.Duration
		has, hasn string
	}{{0, "", first}, {700 * time.Millisecond, first, top}, {2 * time.Second, top, "words"}} {
		u.openedAt = time.Now().Add(-tc.at)
		if s := screen(80, 24); !strings.Contains(s, tc.has) || strings.Contains(s, tc.hasn) {
			t.Errorf("at %v: want %q and not %q", tc.at, tc.has, tc.hasn)
		}
	}
	if s := screen(logoWidth(), 24); strings.Contains(s, first) || !strings.Contains(s, "gotypes") {
		t.Error("too narrow for the logo: the name is typed instead")
	}
	u.key(tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone))
	if s := screen(80, 24); !strings.Contains(s, "words") || u.e.Started {
		t.Fatal("a key should skip the splash without typing")
	}
}

func TestSplashStyles(t *testing.T) {
	defer func(s string) { splashStyle = s }(splashStyle)
	scr := tcell.NewSimulationScreen("UTF-8")
	if err := scr.Init(); err != nil {
		t.Fatal(err)
	}
	scr.SetSize(80, 24)
	row := func(y int) string {
		scr.Show()
		cells, w, _ := scr.GetContents()
		var b strings.Builder
		for _, c := range cells[y*w : (y+1)*w] {
			if len(c.Runes) > 0 {
				b.WriteRune(c.Runes[0])
			}
		}
		return b.String()
	}
	c := canvas{scr, themes[0]}
	for _, splashStyle = range []string{"", "drop", "typo", "scramble", "wave", "press", "fly"} {
		for at := time.Duration(0); at < splashTime; at += 30 * time.Millisecond {
			c.splash(80, 24, at, 5) // every frame draws without going out of range
		}
	}
	splashStyle = "typo"
	c.fill(0, 0, 80, 24, c.st(c.th.text))
	c.splash(80, 24, keyAt(3), 5)
	if r := row(11); !strings.Contains(r, "│ t ││ p │") {
		t.Fatalf("typo: p typed in y's place: %q", r)
	}
}
