package widget

import (
	"github.com/rivo/tview"
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
// it include tview TextView struct
type StatsWidget struct {
	*tview.Flex
	//textbox [...]*tview.TextView
	wpm  *tview.TextView
	acc  *tview.TextView
	cnt  *tview.TextView
	pic  keyboard
	stat stats
}

// ApplyColor set color
func (t *StatsWidget) ApplyColor(p palette) {
	t.SetBackgroundColor(p.background)
	t.wpm.SetBackgroundColor(p.background)
	t.acc.SetBackgroundColor(p.background)
	t.cnt.SetBackgroundColor(p.background)
	t.SetTitleColor(p.title)
	t.SetBorderColor(p.border)
}

func NewStatusWidget() *StatsWidget {
	d := &StatsWidget{
		Flex: tview.NewFlex(),
		wpm:  tview.NewTextView().SetText("WPM : ").SetTextAlign(tview.AlignLeft),
		acc:  tview.NewTextView().SetText("ACCURACY : ").SetTextAlign(tview.AlignLeft),
		cnt:  tview.NewTextView().SetText("COUNT : ").SetTextAlign(tview.AlignLeft),
	}

	d.Flex.SetDirection(tview.FlexRow).
		AddItem(d.wpm, 0, 1, false).
		AddItem(d.acc, 0, 1, false).
		AddItem(d.cnt, 0, 1, false)

	d.SetBorder(true)
	d.SetTitle("Stats")

	return d
}
