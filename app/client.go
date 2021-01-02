package app

import (
	"time"

	"github.com/rivo/tview"
)

var app tview.Application

const corporate = `Leverage agile frameworks to provide a robust synopsis for high level overviews. Iterative approaches to corporate strategy foster collaborative thinking to further the overall value proposition. Organically grow the holistic world view of disruptive innovation via workplace diversity and empowerment.

Bring to the table win-win survival strategies to ensure proactive domination. At the end of the day, going forward, a new normal that has evolved from generation X is on the runway heading towards a streamlined cloud solution. User generated content in real-time will have multiple touchpoints for offshoring.

Capitalize on low hanging fruit to identify a ballpark value added activity to beta test. Override the digital divide with additional clickthroughs from DevOps. Nanotechnology immersion along the information highway will close the loop on focusing solely on the bottom line.

[yellow]Press any key to start`

//--------------------------------------------------------------------

// ThemeList is a box which display text theme list
type ThemeList struct {
	listBox *tview.List
	themes  []string
	current string
}

func NewThemeList() *ThemeList {
	return &ThemeList{
		listBox: tview.NewList(),
		themes:  themes,
	}
}

// TypingBox is a box which display words be typed
type TypingBox struct {
	tview.TextView
	colors Palette
	buffer []string
}

// Dashboard is a frame which display general typing information ( wpm, time ,,)
type Dashboard struct {
	tview.Box
	colors  Palette
	wpm     int
	eplased time.Time
}

//--------------------------------------------------------------------

// Setup general frame layout
func Setup() *tview.Flex {
	pb := NewThemeList()

	flex := tview.NewFlex().
		AddItem(pb.listBox, 0, 1, true)

	return flex
}

func Run(f *tview.Flex) {

	app := tview.NewApplication()

	if err := app.SetRoot(f, true).SetFocus(f).Run(); err != nil {
		panic(err)
	}

	//typing := tview.NewTextView().
	//  SetDynamicColors(true).
	//  SetRegions(true).
	//  SetTextAlign(tview.AlignCenter).
	//  SetChangedFunc(func() {
	//    app.Draw()
	//  })
	//fmt.Fprintf(typing, "%s ", corporate)
	//typing.SetBorder(true)

	//newc := tcell.GetColor("#50fa7b")
	//typing.SetBorderColor(newc)
	//typing.SetTextColor(tcell.GetColor("#bd93f9"))

	//palette := tview.NewBox().
	//  SetBorder(true).
	//  SetBorderAttributes(tcell.AttrBold).
	//  SetTitle("Palette").
	//  SetBackgroundColor(tcell.ColorBlack)

	//grid := tview.NewGrid().
	//  SetRows(1).
	//  SetColumns(20, 0).
	//  SetBorders(false).
	//  SetBordersColor(tcell.ColorBlueViolet)

	//// Layout for screens narrower than 100 cells (menu and side bar are hidden).
	//grid.
	//  AddItem(palette, 0, 0, 0, 0, 0, 0, false).
	//  AddItem(typing, 1, 0, 1, 3, 0, 0, false)
	//// Layout for screens wider than 100 cells.
	//grid.
	//  AddItem(palette, 1, 0, 1, 1, 0, 100, false).
	//  AddItem(typing, 1, 1, 1, 1, 0, 100, false)

	//if err := tview.NewApplication().SetRoot(grid, true).EnableMouse(true).Run(); err != nil {
	//  panic(err)
	//}
}
