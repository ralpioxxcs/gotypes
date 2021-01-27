package app

import (
	"fmt"
	//"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/ralpioxxcs/gotypes/app/widget"
	"github.com/rivo/tview"
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
	a.typing.Input.SetChangedFunc(startTyping)
	//a.typing.Input.SetDoneFunc(diffText)
	//a.typing.Input.SetFinishedFunc(finishtype)

	// set typing frame
	a.flex.AddItem(a.sidebar, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(a.typing, 0, 8, true).
			AddItem(a.stats, 0, 1, false), 0, 9, true)
	a.menuAction(widget.MenuActionImportTheme)

	a.SetRoot(a.flex, true)
	a.EnableMouse(true)

	core = a

	return a
}

// App instance handler
var core *App

// startTyping process typing functions
// * text : current text
func startTyping(text string) {

	/*
	* store start time & elapsed
	* compare current words with indicating words
	*
	 */

	if !core.stats.IsStarted() {
		core.stats.Init(core.typing.GetSentence())

		go func() {
			// set AFK timeout (60 seconds) & update stats each 100 miliseconds tick
			timeout := time.After(60 * time.Second)
			ticker := time.NewTicker(time.Millisecond * 50)
			for range ticker.C {
				select {
				case <-timeout:
					return
				default:
					// Update text stats
					core.QueueUpdateDraw(func() {
						core.stats.Wpm.SetText(fmt.Sprintf("Wpm : %.0f", core.stats.GetNetWpm()))
						core.stats.Accuracy.SetText(fmt.Sprintf("Accuracy : %d", core.stats.GetAccuracy()))
						core.stats.Timer.SetText(fmt.Sprintf("Time : %.02f sec", core.stats.GetElapsed()))
						core.stats.Count.SetText(fmt.Sprintf("Count : %d", core.stats.GetCount()))
					})
				}
			}
		}()
	}

	// compare typing word with target word & coloring , underlining
	core.stats.Stats.Entries = len(text)
	core.typing.Text.SetText("\n\n" + diff(text, core.stats.Stats.Sentence) + "\n\n")

	// compare & check text length
	if len(core.stats.Stats.Sentence) == len(text) {
	}

}

// diffText handles each event keys
func diffText(key tcell.Key) {
	if key == tcell.KeyEnter {
		Logger.Println("enter")
	} else if key == tcell.KeyBackspace {
		Logger.Println("backspace")
	}
}

// diff returns colored string which compareed
func diff(curr string, target string) (colored string) {
	colored = ""

	for i := range curr {
		if curr[i] == target[i] {
			colored += "[green]" + string(curr[i])
		} else {
			colored += "[red]" + string(target[i])
		}
	}
	//colored += "[-:-:-]"
	colored += "[-]"
	for i := len(curr); i < len(target); i++ {
		colored += string(target[i])
	}

	return
}
