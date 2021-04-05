package widget

import (
	"github.com/rivo/tview"
	"time"
)

// keyboard
type keyboard struct {
	temp string
}

type Word struct {
	Text    string
	Iswrong []bool
}

func (w *Word) CopyWord(src string) {
	w.Text = src
	w.Iswrong = make([]bool, len(src))
	for i := range w.Iswrong {
		w.Iswrong[i] = false
	}
}

// Status describe general typing status (wpm, time, accuracy ..)
type Status struct {
	Entries      int       // total character count
	WrongEntries int       // wrong character count
	Words        []Word    // word list
	StartTime    time.Time // start time
	wpm          float64   // words per minute value for display
	accuracy     int       // typing accuracy for display
	count        int       // total typing sentence count
}

func (t *Status) AddCount() {
	t.count += 1
}

func (t *Status) GetCurrentWord() Word {
	return t.Words[t.count-1]
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

func (t *StatusWidget) Reset() {
	t.Status.Entries = 0
	t.Status.WrongEntries = 0
	t.Status.Words = nil
	t.Status.wpm = 0
	t.Status.accuracy = 0
	t.Status.count = 0
	t.start = false
}

// Init initialize string slice of status widget
func (w *StatusWidget) Init(words []string) {
	if w.Status.StartTime.IsZero() {
		w.Status.StartTime = time.Now()
	}

	w.Status.Words = make([]Word, len(words))
	for i, v := range words {
		w.Status.Words[i].CopyWord(v)

		w.Status.Entries += len(v)
	}

	w.start = true
}

// IsStarted returns typing is started
func (t *StatusWidget) IsStarted() bool {
	return t.start
}

// GetGrossWpm returns current wpm (word per minutes)
// * Gross WPM = (All typed entries) / Time (min)
func (t *StatusWidget) GetGrossWpm() float64 {
	return float64(t.Status.count) / time.Since(t.Status.StartTime).Minutes()
}

// GetNetWpm returns current wpm include errors
// * Net WPM = (All typed entries / 5) - ( Uncorrected Errors / Time (min) )
func (t *StatusWidget) GetNetWpm() float64 {
	return t.GetGrossWpm() - (float64(t.Status.WrongEntries) / time.Since(t.Status.StartTime).Minutes())
}

// GetAccuracy returns current word accuracy
func (t *StatusWidget) GetAccuracy() float64 {
	return float64(float64(t.Status.Entries-t.Status.WrongEntries)/float64(t.Status.Entries)) * 100
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
		Entries:      0,
		WrongEntries: 0,
		wpm:          0,
		accuracy:     0,
		count:        1,
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
		AddItem(d.Wpm, 0, 1, false).
		AddItem(d.Accuracy, 0, 1, false).
		AddItem(d.Timer, 0, 1, false).
		AddItem(d.Count, 0, 1, false)

	d.SetBorder(true)
	d.SetTitle("Status")
	return d
}
