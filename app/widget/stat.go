package widget

import (
	"github.com/rivo/tview"
	"strings"
	"time"
)

//
type keyboard struct {
	temp string
}

// Status describe general Status (wpm, time, accuracy ..)
type Status struct {
	Entries    int       // total character count
	Sentence   string    // whole sentence
	WrongCount int       // wrong character count
	AmiWrong   []bool    // if each character is typed wronly for calculate accuracy
	Words      []string  // store each words string array splitted by sentence string
	StartTime  time.Time // start time
	wpm        float64   // words per minute value for display
	accuracy   int       // typing accuracy for display
	count      int       // total typing sentence count
}

func (t *Status) AddCount() {
	t.count += 1
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
	t.SetTitleColor(p.title)
	t.SetBorderColor(p.border)
}

// Init initialize Status if timer is set already just return
func (t *StatusWidget) Init(sentence string) {
	if t.Status.StartTime.IsZero() {
		t.Status.StartTime = time.Now()
	}

	// split sentence into words
	t.Status.Words = strings.Split(sentence, " ")
	t.Status.Sentence = sentence
	t.Status.AmiWrong = make([]bool, len(sentence))
	for i := range sentence {
		t.Status.AmiWrong[i] = false
	}
	t.start = true
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
		Sentence:   "",
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
