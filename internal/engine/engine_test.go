package engine

import (
	"testing"
	"time"
)

type clock struct{ t time.Time }

func (c *clock) now() time.Time      { return c.t }
func (c *clock) add(d time.Duration) { c.t = c.t.Add(d) }

func source(words ...string) func() string {
	i := 0
	return func() string { w := words[i%len(words)]; i++; return w }
}

func newWords(c *clock, words ...string) *Engine {
	return New(Mode{N: len(words)}, source(words...), c.now)
}

func typeStr(e *Engine, s string) {
	for _, r := range s {
		e.Type(r)
	}
}

func TestClockStartsOnFirstKey(t *testing.T) {
	c := &clock{}
	e := newWords(c, "ab", "cd")
	c.add(5 * time.Second)
	e.Type(' ') // space on an empty word is ignored
	if e.Started {
		t.Fatal("started before first real key")
	}
	typeStr(e, "a")
	c.add(2 * time.Second)
	if got := e.Elapsed(); got != 2*time.Second {
		t.Fatalf("elapsed = %v, want 2s", got)
	}
}

func TestSpaceSkipsIncompleteWord(t *testing.T) {
	e := newWords(&clock{}, "ab", "cd")
	typeStr(e, "a ")
	if e.Cur != 1 || e.Words[0].Correct() || e.Words[0].State(1) != Untyped {
		t.Fatalf("cur=%d word0=%+v", e.Cur, e.Words[0])
	}
	if c, i, x := e.Counts(); c != 1 || i != 1 || x != 0 {
		t.Fatalf("counts = %d/%d/%d, want 1/1/0", c, i, x)
	}
}

func TestExtraChars(t *testing.T) {
	e := newWords(&clock{}, "ab", "cd")
	typeStr(e, "abx")
	if e.Words[0].State(2) != Extra {
		t.Fatal("want Extra")
	}
	if _, _, x := e.Counts(); x != 1 {
		t.Fatalf("extra = %d", x)
	}
}

func TestBackspaceOnlyIntoIncorrectWord(t *testing.T) {
	e := newWords(&clock{}, "ab", "cd", "ef")
	typeStr(e, "ab ")
	e.Backspace()
	if e.Cur != 1 {
		t.Fatal("stepped back into a correct word")
	}
	typeStr(e, "cx ")
	e.Backspace()
	e.Backspace()
	if e.Cur != 1 || string(e.Words[1].Typed) != "c" {
		t.Fatalf("cur=%d typed=%q", e.Cur, string(e.Words[1].Typed))
	}
}

func TestDeleteWord(t *testing.T) {
	e := newWords(&clock{}, "abc", "d")
	typeStr(e, "abx")
	e.DeleteWord()
	if len(e.Words[0].Typed) != 0 {
		t.Fatal("word not cleared")
	}
}

func TestCorrectionKeepsError(t *testing.T) {
	e := newWords(&clock{}, "ab", "cd")
	typeStr(e, "x")
	e.Backspace()
	typeStr(e, "a")
	if e.Accuracy() != 50 {
		t.Fatalf("accuracy = %v, want 50", e.Accuracy())
	}
}

func TestWordsModeFinish(t *testing.T) {
	e := newWords(&clock{}, "ab", "cd")
	typeStr(e, "ab cd")
	if !e.Finished {
		t.Fatal("correct last word should finish without space")
	}

	e = newWords(&clock{}, "ab", "cd")
	typeStr(e, "ab cx")
	if e.Finished {
		t.Fatal("finished on a wrong last word")
	}
	typeStr(e, " zz")
	if !e.Finished || e.Cur != 1 || string(e.Words[1].Typed) != "cx" {
		t.Fatalf("space on last word should finish and freeze input: %+v", e.Words[1])
	}
}

func TestWPMAndRaw(t *testing.T) {
	c := &clock{}
	e := newWords(c, "ab", "cd")
	typeStr(e, "ax ")
	c.add(12 * time.Second)
	typeStr(e, "cd")
	// 0.2 min. wpm: only "cd" (2 chars). raw: 5 keystrokes.
	if e.WPM() != 2 || e.Raw() != 5 {
		t.Fatalf("wpm=%v raw=%v, want 2 and 5", e.WPM(), e.Raw())
	}
}

func TestTimeMode(t *testing.T) {
	c := &clock{}
	e := New(Mode{Time: true, N: 10}, source("ab"), c.now)
	for range 100 {
		typeStr(e, "ab ")
	}
	if len(e.Words)-e.Cur < 30 {
		t.Fatalf("ran low on words: %d left", len(e.Words)-e.Cur)
	}
	c.add(10500 * time.Millisecond)
	e.Tick()
	if !e.Finished || e.Elapsed() != 10*time.Second {
		t.Fatalf("finished=%v elapsed=%v", e.Finished, e.Elapsed())
	}
}

func TestRunes(t *testing.T) {
	e := newWords(&clock{}, "가나", "다")
	typeStr(e, "가다라")
	w := e.Words[0]
	if w.State(0) != Correct || w.State(1) != Incorrect || w.State(2) != Extra {
		t.Fatalf("states = %v %v %v", w.State(0), w.State(1), w.State(2))
	}
}

func TestPerSecond(t *testing.T) {
	c := &clock{}
	e := newWords(c, "ab")
	typeStr(e, "a")
	c.add(time.Second)
	typeStr(e, "x")
	e.Backspace()
	c.add(500 * time.Millisecond)
	typeStr(e, "b") // finishes at 1.5s

	ps := e.PerSecond()
	if len(ps) != 2 {
		t.Fatalf("len = %d, want 2", len(ps))
	}
	// second 0: 1 key → 12; second 1: 2 keys in half a second → 48
	if ps[0].Raw != 12 || ps[1].Raw != 48 || ps[0].Errors != 0 || ps[1].Errors != 1 {
		t.Fatalf("per second = %+v", ps)
	}
	if ps[1].WPM != 16 { // 2 chars in 1.5s
		t.Fatalf("wpm = %v, want 16", ps[1].WPM)
	}
}

func TestConsistencySteadyIs100(t *testing.T) {
	c := &clock{}
	e := newWords(c, "ab")
	typeStr(e, "a")
	c.add(2 * time.Second)
	typeStr(e, "b") // sec 2 → 3 seconds sampled: 1, 0, 1 keys
	if got := e.Consistency(); got >= 100 {
		t.Fatalf("uneven typing scored %v", got)
	}

	c = &clock{}
	e = newWords(c, "ab", "cd")
	for i, r := range "ab cd" { // one key per second, last one at exactly 4s
		if i > 0 {
			c.add(time.Second)
		}
		e.Type(r)
	}
	if n := len(e.PerSecond()); n != 5 {
		t.Fatalf("seconds = %d, want 5", n)
	}
	if got := e.Consistency(); got != 100 {
		t.Fatalf("steady typing scored %v", got)
	}
}
