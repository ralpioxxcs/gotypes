package app

import (
	"math"
	"testing"
	"time"

	"github.com/ralpioxxcs/gotypes/internal/engine"
)

func TestGaugeEasesHoldsPeakAndRescales(t *testing.T) {
	now := time.Unix(0, 0)
	u := &ui{lang: "english", lists: map[string][]string{"english": {"a"}}, mode: 1}
	u.restart()
	u.e = engine.New(engine.Mode{N: 1000}, func() string { return "a" }, func() time.Time { return now })
	frame := func(d time.Duration) {
		for end := now.Add(d); now.Before(end); now = now.Add(33 * time.Millisecond) {
			u.animate(now)
		}
	}
	// 20 keys/s ("a " pairs) = 1200 keys/min = 240 wpm, above the 150 default scale
	for range 40 {
		u.e.Type('a')
		frame(50 * time.Millisecond)
		u.e.Type(' ')
		frame(50 * time.Millisecond)
	}
	if want := u.speed(); math.Abs(u.gauge.shown-want) > want*0.05 {
		t.Fatalf("shown %.0f did not converge to %.0f", u.gauge.shown, want)
	}
	if u.gauge.scaleTo < u.gauge.peak {
		t.Fatalf("scale %.0f not grown past peak %.0f", u.gauge.scaleTo, u.gauge.peak)
	}
	peak := u.gauge.peak
	frame(800 * time.Millisecond) // stopped typing: value falls, peak holds
	if u.gauge.shown >= peak || u.gauge.peak != peak {
		t.Fatalf("after 0.8s idle: shown %.0f peak %.0f (was %.0f)", u.gauge.shown, u.gauge.peak, peak)
	}
	frame(time.Second)
	if u.gauge.peak >= peak {
		t.Fatal("peak should start falling after holding ~1s")
	}
}
