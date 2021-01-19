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
	CurrentWord string
	Words       []string
	StartTime   time.Time
	wpm         string
	accuracy    int
	count       int
}

// StatsWidget is frame which display general typing information ( wpm, time ,,)
// it include tview.TextView struct
type StatsWidget struct {
	*tview.Flex
	Wpm     *tview.TextView
	Accuray *tview.TextView
	Timer   *tview.TextView
	Count   *tview.TextView
	pic     keyboard
	Stats   stats
}

// ApplyColor set color
func (t *StatsWidget) ApplyColor(p palette) {
	t.SetBackgroundColor(p.background)
	t.Wpm.SetBackgroundColor(p.background)
	t.Accuray.SetBackgroundColor(p.background)
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
}

func (t *StatsWidget) GetWpm() float64 {
	return 1.2
}

func (t *StatsWidget) GetElapsed() float64 {
	return time.Since(t.Stats.StartTime).Seconds()
}

// NewStatusWidget returns initialized StatsWidget
func NewStatusWidget() *StatsWidget {
	d := &StatsWidget{
		Flex:    tview.NewFlex(),
		Wpm:     tview.NewTextView().SetText("WPM : ").SetTextAlign(tview.AlignLeft),
		Accuray: tview.NewTextView().SetText("ACCURACY : ").SetTextAlign(tview.AlignLeft),
		Timer:   tview.NewTextView().SetText("TIME : ").SetTextAlign(tview.AlignLeft),
		Count:   tview.NewTextView().SetText("COUNT : ").SetTextAlign(tview.AlignLeft),
	}

	d.Flex.SetDirection(tview.FlexRow).
		AddItem(d.Wpm, 0, 1, false).
		AddItem(d.Accuray, 0, 1, false).
		AddItem(d.Timer, 0, 1, false).
		AddItem(d.Count, 0, 1, false)

	d.SetBorder(true)
	d.SetTitle("Stats")

	return d
}
