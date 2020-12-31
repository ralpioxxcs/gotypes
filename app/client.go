package app

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

func Run() {

	// primitive type based on text view
	newPrimitive := func(text string) tview.Primitive {
		return tview.NewTextView().
			SetTextAlign(tview.AlignCenter).
			SetText("").
			SetTextColor(tcell.ColorBlue).
			SetBackgroundColor(tcell.ColorWheat)
	}

	main := newPrimitive("Main")
	box := tview.NewBox().
		SetBorder(true).
		SetBorderAttributes(tcell.AttrBold).
		SetTitle("Palette").
		SetBackgroundColor(tcell.ColorWheat)

	grid := tview.NewGrid().
		SetRows(1).
		SetColumns(20, 0).
		SetBorders(false).
		SetBordersColor(tcell.ColorBlueViolet)

	//grid.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
	//grid.SetBackgroundColor(tview.Styles.ContrastBackgroundColor)

	// Layout for screens narrower than 100 cells (menu and side bar are hidden).
	grid.AddItem(box, 0, 0, 0, 0, 0, 0, false).
		AddItem(main, 1, 0, 1, 3, 0, 0, false)

	// Layout for screens wider than 100 cells.
	grid.AddItem(box, 1, 0, 1, 1, 0, 100, false).
		AddItem(main, 1, 1, 1, 1, 0, 100, false)

	if err := tview.NewApplication().SetRoot(grid, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
