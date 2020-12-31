package app

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

func Run() {
	newPrimitive := func(text string) tview.Primitive {
		return tview.NewTextView().
			SetTextAlign(tview.AlignCenter).
			SetText(text)
		//SetBackgroundColor(tcell.ColorRoyalBlue)
	}
	palette := newPrimitive("Palette")
	main := newPrimitive("Main")

	grid := tview.NewGrid().
		SetRows(1).
		SetColumns(15, 0).
		SetBorders(true).
		SetBordersColor(tcell.ColorBlueViolet)

	grid.SetTitleColor(tcell.ColorGold)

	//grid.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
	//grid.SetBackgroundColor(tview.Styles.ContrastBackgroundColor)

	// Layout for screens narrower than 100 cells (menu and side bar are hidden).
	grid.AddItem(palette, 0, 0, 0, 0, 0, 0, false).
		AddItem(main, 1, 0, 1, 3, 0, 0, false)

	// Layout for screens wider than 100 cells.
	grid.AddItem(palette, 1, 0, 1, 1, 0, 100, false).
		AddItem(main, 1, 1, 1, 1, 0, 100, false)

	if err := tview.NewApplication().SetRoot(grid, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
