// Package engine is the typing test logic, free of any UI.
// Rules follow monkeytype: the first key starts the clock, space moves to the
// next word, corrections never erase recorded errors.
package engine

import (
	"math"
	"slices"
	"time"
)

// Mode is N words, or N seconds when Time is set.
type Mode struct {
	Time bool
	N    int
}

type CharState int

const (
	Untyped CharState = iota
	Correct
	Incorrect
	Extra // typed past the end of the word
)

type Word struct {
	Target []rune
	Typed  []rune
}

func (w Word) State(i int) CharState {
	switch {
	case i >= len(w.Typed):
		return Untyped
	case i >= len(w.Target):
		return Extra
	case w.Typed[i] == w.Target[i]:
		return Correct
	}
	return Incorrect
}

func (w Word) Correct() bool { return slices.Equal(w.Typed, w.Target) }

// Second is one point of the per-second result chart.
type Second struct {
	WPM    float64 // cumulative wpm at the end of this second
	Raw    float64 // raw speed within this second only
	Errors int
}

// Engine fields are read by the UI; change state only through methods.
type Engine struct {
	Mode     Mode
	Words    []Word
	Cur      int
	Started  bool
	Finished bool

	next       func() string
	now        func() time.Time
	start, end time.Time

	correct, incorrect, extra int // keystrokes
	keysAt, errAt             []int
	wpmAt                     []float64
	hits                      []hit // correct keys, for Recent
}

type hit struct {
	t time.Time
	r rune
}

// New starts a test. next supplies random words; now is the clock (time.Now outside tests).
func New(mode Mode, next func() string, now func() time.Time) *Engine {
	e := &Engine{Mode: mode, next: next, now: now}
	n := mode.N
	if mode.Time {
		n = 0 // topUp fills
	}
	e.add(n)
	e.topUp()
	return e
}

func (e *Engine) add(n int) {
	for range n {
		e.Words = append(e.Words, Word{Target: []rune(e.next())})
	}
}

// topUp keeps time mode from running out of words.
func (e *Engine) topUp() {
	if e.Mode.Time && len(e.Words)-e.Cur < 30 {
		e.add(50)
	}
}

func (e *Engine) Type(r rune) {
	if e.Finished {
		return
	}
	w := &e.Words[e.Cur]
	if r == ' ' {
		if len(w.Typed) == 0 {
			return
		}
		e.key(w.Correct(), false)
		e.hit(w.Correct(), ' ')
		if !e.Mode.Time && e.Cur == len(e.Words)-1 {
			e.finish()
			return
		}
		e.Cur++
		e.topUp()
		return
	}
	i := len(w.Typed)
	w.Typed = append(w.Typed, r)
	e.key(w.State(i) == Correct, w.State(i) == Extra)
	e.hit(w.State(i) == Correct, r)
	if !e.Mode.Time && e.Cur == len(e.Words)-1 && w.Correct() {
		e.finish()
	}
}

func (e *Engine) key(ok, extra bool) {
	if !e.Started {
		e.Started = true
		e.start = e.now()
	}
	sec := e.second()
	e.keysAt = bump(e.keysAt, sec)
	switch {
	case ok:
		e.correct++
		return
	case extra:
		e.extra++
	default:
		e.incorrect++
	}
	e.errAt = bump(e.errAt, sec)
}

func (e *Engine) hit(ok bool, r rune) {
	if ok {
		e.hits = append(e.hits, hit{e.now(), r})
	}
}

// Recent is the live speed: weighted correct keys per minute over the last
// window (at least 1s, so the first keys don't spike). Use weight 1 and /5
// for wpm, Strokes for 타/분.
func (e *Engine) Recent(window time.Duration, weight func(rune) int) float64 {
	if !e.Started {
		return 0
	}
	now := e.now()
	if e.Finished {
		now = e.end
	}
	span := min(window, max(now.Sub(e.start), time.Second))
	n := 0
	for i := len(e.hits) - 1; i >= 0 && now.Sub(e.hits[i].t) < span; i-- {
		n += weight(e.hits[i].r)
	}
	return float64(n) / span.Minutes()
}

func bump(s []int, i int) []int {
	for len(s) <= i {
		s = append(s, 0)
	}
	s[i]++
	return s
}

func (e *Engine) Backspace()  { e.erase(false) }
func (e *Engine) DeleteWord() { e.erase(true) }

func (e *Engine) erase(word bool) {
	if e.Finished {
		return
	}
	w := &e.Words[e.Cur]
	if len(w.Typed) == 0 {
		// only an incorrect word may be revisited
		if e.Cur > 0 && !e.Words[e.Cur-1].Correct() {
			e.Cur--
		}
		return
	}
	if word {
		w.Typed = w.Typed[:0]
	} else {
		w.Typed = w.Typed[:len(w.Typed)-1]
	}
}

// Tick ends time mode and records per-second wpm. Call it a few times a second.
func (e *Engine) Tick() {
	if !e.Started || e.Finished {
		return
	}
	if e.Mode.Time && e.Elapsed() >= e.limit() {
		e.finish()
		return
	}
	e.sample(e.second())
}

func (e *Engine) limit() time.Duration { return time.Duration(e.Mode.N) * time.Second }

func (e *Engine) second() int { return int(e.now().Sub(e.start) / time.Second) }

// sample fills wpm up to n seconds.
// shortcut: seconds missed between Tick calls get the wpm of the late tick, fine at UI tick rates.
func (e *Engine) sample(n int) {
	for len(e.wpmAt) < n {
		e.wpmAt = append(e.wpmAt, e.WPM())
	}
}

func (e *Engine) finish() {
	e.Finished = true
	e.end = e.now()
	if e.Mode.Time && e.end.Sub(e.start) > e.limit() {
		e.end = e.start.Add(e.limit())
	}
	// a key landing exactly on a whole second opens that second
	e.sample(max(int(math.Ceil(e.Elapsed().Seconds())), len(e.keysAt)))
}

func (e *Engine) Elapsed() time.Duration {
	switch {
	case !e.Started:
		return 0
	case e.Finished:
		return e.end.Sub(e.start)
	}
	return e.now().Sub(e.start)
}

func (e *Engine) perMinute(chars int) float64 {
	m := e.Elapsed().Minutes()
	if m == 0 {
		return 0
	}
	return float64(chars) / 5 / m
}

// correctUnits weighs the chars of correctly typed words, plus 1 per space after each.
func (e *Engine) correctUnits(weight func(rune) int) int {
	n := 0
	for i := 0; i <= e.Cur; i++ {
		if w := e.Words[i]; w.Correct() {
			for _, r := range w.Target {
				n += weight(r)
			}
			if i < e.Cur {
				n++
			}
		}
	}
	return n
}

// WPM counts only correctly typed words, plus the space after each.
func (e *Engine) WPM() float64 { return e.perMinute(e.correctUnits(func(rune) int { return 1 })) }

// Taja is Korean-style speed (타/분): keystrokes of correctly typed words per minute.
func (e *Engine) Taja() float64 { return e.perMinute(e.correctUnits(Strokes)) * 5 }

// Strokes is how many keys r takes on a 2-set (두벌식) Korean keyboard; 1 for non-Hangul.
// Doubled consonants and ㅒ/ㅖ are one shifted key; compound vowels and finals are two.
func Strokes(r rune) int {
	if r < 0xAC00 || r > 0xD7A3 {
		return 1
	}
	s := int(r - 0xAC00)
	n := 2 // initial + vowel
	switch s % 588 / 28 {
	case 9, 10, 11, 14, 15, 16, 19: // ㅘㅙㅚㅝㅞㅟㅢ
		n++
	}
	switch s % 28 {
	case 0:
	case 3, 5, 6, 9, 10, 11, 12, 13, 14, 15, 18: // ㄳㄵㄶㄺㄻㄼㄽㄾㄿㅀㅄ
		n += 2
	default:
		n++
	}
	return n
}

// Raw counts every keystroke.
func (e *Engine) Raw() float64 { return e.perMinute(e.correct + e.incorrect + e.extra) }

func (e *Engine) Accuracy() float64 {
	total := e.correct + e.incorrect + e.extra
	if total == 0 {
		return 100
	}
	return float64(e.correct) * 100 / float64(total)
}

// Counts returns keystrokes: correct, incorrect, extra.
func (e *Engine) Counts() (correct, incorrect, extra int) {
	return e.correct, e.incorrect, e.extra
}

func (e *Engine) PerSecond() []Second {
	out := make([]Second, len(e.wpmAt))
	el := e.Elapsed().Seconds()
	for i := range out {
		out[i].WPM = e.wpmAt[i]
		if i < len(e.keysAt) {
			out[i].Raw = float64(e.keysAt[i]) * 12 // keys/sec → wpm
			if f := el - float64(i); f > 0 && f < 1 {
				out[i].Raw /= f // last, partial second
			}
		}
		if i < len(e.errAt) {
			out[i].Errors = e.errAt[i]
		}
	}
	return out
}

// Consistency is 100 minus the coefficient of variation (%) of per-second raw speed.
func (e *Engine) Consistency() float64 {
	ps := e.PerSecond()
	if len(ps) < 2 {
		return 100
	}
	var sum, sq float64
	for _, s := range ps {
		sum += s.Raw
	}
	mean := sum / float64(len(ps))
	if mean == 0 {
		return 0
	}
	for _, s := range ps {
		sq += (s.Raw - mean) * (s.Raw - mean)
	}
	return math.Max(0, 100*(1-math.Sqrt(sq/float64(len(ps)))/mean))
}
