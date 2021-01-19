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
	Wpm   *tview.TextView
	Acc   *tview.TextView
	Cnt   *tview.TextView
	pic   keyboard
	Stats stats
}

// ApplyColor set color
func (t *StatsWidget) ApplyColor(p palette) {
	t.SetBackgroundColor(p.background)
	t.Wpm.SetBackgroundColor(p.background)
	t.Acc.SetBackgroundColor(p.background)
	t.Cnt.SetBackgroundColor(p.background)
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

func (t *StatsWidget) GetElapsed() int64 {
	return time.Since(t.Stats.StartTime).Milliseconds()
}

// NewStatusWidget returns initialized StatsWidget
func NewStatusWidget() *StatsWidget {
	d := &StatsWidget{
		Flex: tview.NewFlex(),
		Wpm:  tview.NewTextView().SetText("WPM : ").SetTextAlign(tview.AlignLeft),
		Acc:  tview.NewTextView().SetText("ACCURACY : ").SetTextAlign(tview.AlignLeft),
		Cnt:  tview.NewTextView().SetText("COUNT : ").SetTextAlign(tview.AlignLeft),
	}

	d.Flex.SetDirection(tview.FlexRow).
		AddItem(d.Wpm, 0, 1, false).
		AddItem(d.Acc, 0, 1, false).
		AddItem(d.Cnt, 0, 1, false)

	d.SetBorder(true)
	d.SetTitle("Stats")

	return d
}
