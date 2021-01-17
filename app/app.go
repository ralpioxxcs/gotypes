package app

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"

	"github.com/ralpioxxcs/gotypes/app/widget"
)

/* App is entire tui struct including tview flex struct
 * it consists of each widgets (sidebar, body, status)
 * - sidebar : it lists color themes
 * - body    : display words and current carrot interactively
 * - stats   : it shows current status such as wpm, time ..
 */
type App struct {
	*tview.Application
	flex    *tview.Flex
	sidebar *widget.ThemeList
	typing  *widget.TypingWidget
	stats   *widget.StatsWidget
}

func (a *App) Draw(screen tcell.Screen) {
	a.flex.Draw(screen)
}

func (a *App) GetRect() (int, int, int, int) {
	return a.flex.GetRect()
}

func (a *App) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		a.flex.InputHandler()(event, setFocus)
	}
}

func (a *App) HasFocus() bool {
	return a.flex.HasFocus()
}

func (a *App) Focus(delegate func(p tview.Primitive)) {
	a.flex.Focus(delegate)
}

func (a *App) Blur() {
	a.flex.Blur()
}

// menuAction apply current theme each widget
func (a *App) menuAction(action widget.MenuAction) {
	switch action {
	case widget.MenuActionNone:
	case widget.MenuActionImportTheme:
		a.sidebar.ApplyColor(a.sidebar.Theme[a.sidebar.GetCurrentTheme()])
		a.typing.ApplyColor(a.sidebar.Theme[a.sidebar.GetCurrentTheme()])
		a.stats.ApplyColor(a.sidebar.Theme[a.sidebar.GetCurrentTheme()])
	}
}

// NewApp returns initialized App struct
func NewApp() *App {
	a := &App{
		Application: tview.NewApplication(),
		flex:        tview.NewFlex(),
		sidebar:     widget.NewThemeList(),
		typing:      widget.NewTypingWidget(),
		stats:       widget.NewStatusWidget(),
	}

	// set function to sidebar
	a.sidebar.SetActionFunc(a.menuAction)

	// typing
	a.typing.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key != nil {
			a.SetFocus(a.typing.Input)
		}
		return event
	})
	a.typing.Input.SetChangedFunc(test)

	// set typing frame
	a.flex.AddItem(a.sidebar, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(a.typing, 0, 8, true).
			AddItem(a.stats, 0, 1, false), 0, 9, true)
	a.menuAction(widget.MenuActionImportTheme)

	a.SetRoot(a.flex, true)
	a.EnableMouse(true)

	return a
}

func test(text string) {
	Logger.Println(text)

}
