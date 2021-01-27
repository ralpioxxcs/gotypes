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
	Entries   int
	Wrong     int
	Sentence  string
	Words     []string
	StartTime time.Time
	wpm       float64
	accuracy  int
	count     int
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
	Stats    *stats
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
	t.Stats.Sentence = sentence
	t.start = true
}

// IsStarted returns typing is started
func (t *StatsWidget) IsStarted() bool {
	return t.start
}

// GetGrossWpm returns current wpm (word per minutes)
// * Gross WPM = (All typed entries / 5) / Time (min)
func (t *StatsWidget) GetGrossWpm() float64 {
	return float64(t.Stats.Entries/5) / time.Since(t.Stats.StartTime).Minutes()
}

// GetNetWpm returns current wpm include errors
// * Net WPM = (All typed entries / 5) - ( Uncorrected Errors / Time (min) )
func (t *StatsWidget) GetNetWpm() float64 {
	return t.GetGrossWpm() - (float64(t.Stats.Wrong) / time.Since(t.Stats.StartTime).Minutes())
}

// GetAccuracy returns current word accuracy
func (t *StatsWidget) GetAccuracy() int {
	if t.GetNetWpm() == 0 {
		return 100
	}
	return (int(t.GetNetWpm()) / int(t.GetGrossWpm())) * 100
}

// GetElapsed returns current time elapsed
func (t *StatsWidget) GetElapsed() float64 {
	return time.Since(t.Stats.StartTime).Seconds()
}

// GetCount returns typed sentence count
func (t *StatsWidget) GetCount() int {
	return t.Stats.count
}

func NewStats() *stats {
	s := &stats{
		Entries:  0,
		Wrong:    0,
		Sentence: "",
		wpm:      0,
		accuracy: 0,
		count:    0,
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
		Stats:    NewStats(),
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
