package widget

import (
	"github.com/rivo/tview"
	"strings"
	"time"
)

type keyboard struct {
	temp string
}

// Stats describe general stats (wpm, time, accuracy ..)
type stats struct {
	Index       int
	CurrentWord string
	Words       []string
	StartTime   time.Time
	wpm         float64
	accuracy    int
	count       int
}

// StatsWidget is frame which display general typing information ( wpm, time ,,)
// it include tview.TextView struct
type StatsWidget struct {
	*tview.Flex
	start    bool
	Wpm      *tview.TextView
	Accuracy *tview.TextView
	Timer    *tview.TextView
	Count    *tview.TextView
	pic      keyboard
	Stats    stats
}

// ApplyColor set color
func (t *StatsWidget) ApplyColor(p palette) {
	t.SetBackgroundColor(p.background)
	t.Wpm.SetBackgroundColor(p.background)
	t.Accuracy.SetBackgroundColor(p.background)
	t.Timer.SetBackgroundColor(p.background)
	t.Count.SetBackgroundColor(p.background)
	t.SetTitleColor(p.title)
	t.SetBorderColor(p.border)
}

// Init initialize stats if timer is set already just return
func (t *StatsWidget) Init(sentence string) {
	if t.Stats.StartTime.IsZero() {
		t.Stats.StartTime = time.Now()
	} else {
		return
	}

	// split sentence into words
	t.Stats.Words = strings.Split(sentence, " ")
	t.Stats.CurrentWord = t.Stats.Words[t.Stats.Index]
	t.start = true
}

func (t *StatsWidget) IsStarted() bool {
	return t.start
}

// GetWpm returns current wpm (word per minutes)
func (t *StatsWidget) GetWpm() float64 {
	return 1.5
}

// GetAccuracy returns current word accuracy
func (t *StatsWidget) GetAccuracy() int {
	return 100
}

// GetElapsed returns current time elapsed
func (t *StatsWidget) GetElapsed() float64 {
	return time.Since(t.Stats.StartTime).Seconds()
}

func NewStats() *stats {
	s := &stats{
		Index:       0,
		CurrentWord: "",
		wpm:         0,
		accuracy:    0,
		count:       0,
	}
	return s
}

// NewStatusWidget returns initialized StatsWidget
func NewStatusWidget() *StatsWidget {
	d := &StatsWidget{
		start:    false,
		Flex:     tview.NewFlex(),
		Wpm:      tview.NewTextView().SetText("Wpm : ").SetTextAlign(tview.AlignLeft),
		Accuracy: tview.NewTextView().SetText("Accuracy : ").SetTextAlign(tview.AlignLeft),
		Timer:    tview.NewTextView().SetText("Time : ").SetTextAlign(tview.AlignLeft),
		Count:    tview.NewTextView().SetText("Count : ").SetTextAlign(tview.AlignLeft),
		Stats:    *NewStats(),
	}

	d.Flex.SetDirection(tview.FlexRow).
		AddItem(d.Wpm, 0, 1, false).
		AddItem(d.Accuracy, 0, 1, false).
		AddItem(d.Timer, 0, 1, false).
		AddItem(d.Count, 0, 1, false)

	d.SetBorder(true)
	d.SetTitle("Stats")

	return d
}
