package widget

import (
	"github.com/rivo/tview"
	_ "strings"
	"time"
)

//
type keyboard struct {
	temp string
}

// Status describe general Status (wpm, time, accuracy ..)
type Status struct {
	Entries    int       // total character count
	WrongCount int       // wrong character count
	AmiWrong   []bool    // if each character is typed wronly for calculate accuracy
	Words      []string  // store each words
	StartTime  time.Time // start time
	wpm        float64   // words per minute value for display
	accuracy   int       // typing accuracy for display
	count      int       // total typing sentence count
}

func (t *Status) AddCount() {
	t.count += 1
}

func (t *Status) GetCurrentWord() string {
	return t.Words[t.count]
}

// StatusWidget is frame which display general typing information ( wpm, time ,,)
// it include tview.TextView struct
type StatusWidget struct {
	*tview.Flex
	Wpm      *tview.TextView
	Accuracy *tview.TextView
	Timer    *tview.TextView
	Count    *tview.TextView
	Status   *Status
	start    bool
	pic      keyboard
}

// ApplyColor set color
func (t *StatusWidget) ApplyColor(p palette) {
	t.SetBackgroundColor(p.background)
	t.Wpm.SetBackgroundColor(p.background)
	t.Accuracy.SetBackgroundColor(p.background)
	t.Timer.SetBackgroundColor(p.background)
	t.Count.SetBackgroundColor(p.background)

	t.Wpm.SetTextColor(p.title)
	t.Accuracy.SetTextColor(p.title)
	t.Timer.SetTextColor(p.title)
	t.Count.SetTextColor(p.title)

	t.SetTitleColor(p.title)
	t.SetBorderColor(p.border)
}

func (w *StatusWidget) Init(words []string) {
	if w.Status.StartTime.IsZero() {
		w.Status.StartTime = time.Now()
	}
	w.Status.Words = words
	w.Status.AmiWrong = make([]bool, len(words))
	for i := range words {
		w.Status.AmiWrong[i] = false
	}
	w.start = true
}

// IsStarted returns typing is started
func (t *StatusWidget) IsStarted() bool {
	return t.start
}

// GetGrossWpm returns current wpm (word per minutes)
// * Gross WPM = (All typed entries / 5) / Time (min)
func (t *StatusWidget) GetGrossWpm() float64 {
	return float64(t.Status.Entries/5) / time.Since(t.Status.StartTime).Minutes()
}

// GetNetWpm returns current wpm include errors
// * Net WPM = (All typed entries / 5) - ( Uncorrected Errors / Time (min) )
func (t *StatusWidget) GetNetWpm() float64 {
	return t.GetGrossWpm() - (float64(t.Status.WrongCount) / time.Since(t.Status.StartTime).Minutes())
}

// GetAccuracy returns current word accuracy
func (t *StatusWidget) GetAccuracy() float64 {
	if t.Status.Entries == 0 {
		return 0
	} else if t.Status.WrongCount == 0 {
		return 100
	}
	return float64(float64(t.Status.Entries-t.Status.WrongCount)/float64(t.Status.Entries)) * 100
}

// GetElapsed returns current time elapsed
func (t *StatusWidget) GetElapsed() float64 {
	return time.Since(t.Status.StartTime).Seconds()
}

// GetCount returns typed sentence count
func (t *StatusWidget) GetCount() int {
	return t.Status.count
}

func NewStatus() *Status {
	s := &Status{
		Entries:    0,
		WrongCount: 0,
		wpm:        0,
		accuracy:   0,
		count:      0,
	}
	return s
}

// NewStatusWidget returns initialized StatusWidget
func NewStatusWidget() *StatusWidget {
	d := &StatusWidget{
		start:    false,
		Flex:     tview.NewFlex(),
		Wpm:      tview.NewTextView().SetText("Wpm : ").SetTextAlign(tview.AlignLeft),
		Accuracy: tview.NewTextView().SetText("Accuracy : ").SetTextAlign(tview.AlignLeft),
		Timer:    tview.NewTextView().SetText("Time : ").SetTextAlign(tview.AlignLeft),
		Count:    tview.NewTextView().SetText("Count : ").SetTextAlign(tview.AlignLeft),
		Status:   NewStatus(),
	}

	d.Flex.SetDirection(tview.FlexRow).
		AddItem(d.Wpm, 1, 0, false).
		AddItem(d.Accuracy, 1, 0, false).
		AddItem(d.Timer, 1, 0, false).
		AddItem(d.Count, 1, 0, false)

	d.SetBorder(true)
	d.SetTitle("Status")
	return d
}
