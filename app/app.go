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
 * it consists of each widgets (sidebarWidget, body, StatsuWidget)
 * - sidebarWidget : it lists color themes
 * - body    : display words and current carrot interactively
 * - StatsuWidget   : it shows current status such as wpm, time ..
 */
type App struct {
	*tview.Application
	flex          *tview.Flex
	sidebarWidget *widget.ThemeList
	typingWidget  *widget.TypingWidget
	statusWidget  *widget.StatusWidget
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
		a.sidebarWidget.ApplyColor(a.sidebarWidget.Theme[a.sidebarWidget.GetCurrentTheme()])
		a.typingWidget.ApplyColor(a.sidebarWidget.Theme[a.sidebarWidget.GetCurrentTheme()])
		a.statusWidget.ApplyColor(a.sidebarWidget.Theme[a.sidebarWidget.GetCurrentTheme()])
	}
}

// NewApp returns initialized App struct
func NewApp() *App {
	a := &App{
		Application:   tview.NewApplication(),
		flex:          tview.NewFlex(),
		sidebarWidget: widget.NewThemeList(),
		typingWidget:  widget.NewTypingWidget(),
		statusWidget:  widget.NewStatusWidget(),
	}

	// set function to sidebarWidget
	a.sidebarWidget.SetActionFunc(a.menuAction)

	// typingWidget
	a.typingWidget.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key != nil {
			a.SetFocus(a.typingWidget.Input)
		}
		return event
	})
	a.typingWidget.Input.SetChangedFunc(startTyping)
	//a.typingWidget.Input.SetDoneFunc(diffText)
	//a.typingWidget.Input.SetFinishedFunc(finishtype)

	// set typingWidget frame
	a.flex.AddItem(a.sidebarWidget, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(a.typingWidget, 0, 8, true).
			AddItem(a.statusWidget, 0, 1, false), 0, 9, true)
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

	if !core.statusWidget.IsStarted() {
		core.statusWidget.Init(core.typingWidget.GetSentence())

		go func() {
			// set AFK timeout (60 seconds) & update statusWidget each 100 miliseconds tick
			timeout := time.After(60 * time.Second)
			ticker := time.NewTicker(time.Millisecond * 50)
			for range ticker.C {
				select {
				case <-timeout:
					return
				default:
					// Update text statusWidget
					core.QueueUpdateDraw(func() {
						core.statusWidget.Wpm.SetText(fmt.Sprintf("Wpm : %.0f", core.statusWidget.GetNetWpm()))
						core.statusWidget.Accuracy.SetText(fmt.Sprintf("Accuracy : %d", core.statusWidget.GetAccuracy()))
						core.statusWidget.Timer.SetText(fmt.Sprintf("Time : %.02f sec", core.statusWidget.GetElapsed()))
						core.statusWidget.Count.SetText(fmt.Sprintf("Count : %d", core.statusWidget.GetCount()))
					})
				}
			}
		}()
	}

	// compare typingWidget word with target word & coloring , underlining
	core.statusWidget.Status.Entries = len(text)
	core.typingWidget.Text.SetText("\n\n" + diff(text, core.statusWidget.Status.Sentence) + "\n\n")

	// check error character

	// compare & check text length
	if len(core.statusWidget.Status.Sentence) == len(text) {
		pages := tview.NewPages().
			AddPage("modal", tview.NewModal().
				SetText("End").
				SetBackgroundColor(tcell.ColorDefault).
				AddButtons([]string{"exit"}).SetDoneFunc(func(index int, label string) {
			}), false, false)
		pages.ShowPage("end")
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
			core.statusWidget.Status.Amiwrong[i] = false
		} else {
			colored += "[red]" + string(target[i])
			if core.statusWidget.Status.Amiwrong[i] == false {
				core.statusWidget.Status.Amiwrong[i] = true
				core.statusWidget.Status.Wrong++
			}
		}
	}
	colored += "[-]"
	for i := len(curr); i < len(target); i++ {
		colored += string(target[i])
	}

	return
}
