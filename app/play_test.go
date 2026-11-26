package app

import (
	"context"
	"runtime"
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
