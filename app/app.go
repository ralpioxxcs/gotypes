package app

import (
	"fmt"
	//"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/ralpioxxcs/gotypes/app/widget"
	"github.com/rivo/tview"
)

/*
 * App is entire tui struct including tview flex struct
 * it consists of each widgets (side-bar, body, status)
 * - side-bar : it lists color themes
 * - body : display words and current carrot interactively
 * - status : it shows current status such as wpm, time ..
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
		bg := a.sidebarWidget.Theme[a.sidebarWidget.GetCurrentTheme()]
		a.flex.SetBackgroundColor(bg.GetBg())
		a.sidebarWidget.ApplyColor(a.sidebarWidget.Theme[a.sidebarWidget.GetCurrentTheme()])
		a.typingWidget.ApplyColor(a.sidebarWidget.Theme[a.sidebarWidget.GetCurrentTheme()])
		a.statusWidget.ApplyColor(a.sidebarWidget.Theme[a.sidebarWidget.GetCurrentTheme()])
	}
}

// Reset state
func (a *App) Reset() {
	//a.statusWidget.Init()
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

	// set function to side-bar widget
	a.sidebarWidget.SetActionFunc(a.menuAction)

	// config callback functions of typing widget
	// -> set focus to typing widget first time
	a.typingWidget.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key != nil {
			a.SetFocus(a.typingWidget.Input)
		}
		return event
	})
	// -> realtime typing process callback
	a.typingWidget.Input.SetChangedFunc(startTyping)
	// -> SetDoneFunc sets a handler which is called when the user is done entering text.
	a.typingWidget.Input.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyBackspace {
			a.statusWidget.Status.AmiWrong[a.statusWidget.Status.Entries+1] = false
		}
		if key == tcell.KeyTab {
			a.Reset()
		}
	})

	// set app layout
	a.flex.AddItem(a.sidebarWidget, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(a.typingWidget, 0, 2, true).
			AddItem(a.statusWidget, 0, 1, false), 0, 10, true)
	a.menuAction(widget.MenuActionImportTheme)

	a.SetRoot(a.flex, true)
	a.EnableMouse(true)

	// config modal popup
	popup = tview.NewModal().
		SetText("Do you want replay?").
		AddButtons([]string{"Yes", "Cancel"}).
		SetBackgroundColor(tcell.ColorDefault).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Yes" {
				a.Stop()
			} else {
			}
		})

	core = a

	return a
}

// App instance handler
var core *App
var popup *tview.Modal

// startTyping process typing functions
// * text : current text
func startTyping(text string) {

	/*
	* store start time & elapsed
	* compare current words with indicating words
	 */
	if !core.statusWidget.IsStarted() {
		core.statusWidget.Init(core.typingWidget.GetSentence())

		go func() {
			// set AFK timeout (60 seconds) & update status widget each 100 miliseconds tick
			timeout := time.After(60 * time.Second)
			ticker := time.NewTicker(time.Millisecond * 50)
			for range ticker.C {
				select {
				case <-timeout:
					return
				default:
					// Update text status widget
					core.QueueUpdateDraw(func() {
						core.statusWidget.Wpm.
							SetText(fmt.Sprintf("Wpm : %.0f", core.statusWidget.GetNetWpm()))
						core.statusWidget.Accuracy.
							SetText(fmt.Sprintf("Accuracy : %.0f", core.statusWidget.GetAccuracy()))
						core.statusWidget.Timer.
							SetText(fmt.Sprintf("Time : %.02f sec", core.statusWidget.GetElapsed()))
						core.statusWidget.Count.
							SetText(fmt.Sprintf("Count : %d", core.statusWidget.GetCount()))
					})
				}
			}
		}()
	}

	// compare typingWidget word with target word & coloring , underlining
	core.statusWidget.Status.Entries = len(text)
	core.typingWidget.Text.
		SetText("\n\n" + diff(text, core.statusWidget.Status.Sentence) + "\n\n")

	// compare & check text length
	if (len(core.statusWidget.Status.Sentence)) == len(text) {
		core.statusWidget.Status.AddCount()
		//core.SetRoot(popup, false).SetFocus(popup).Run()

		// set next sentence
		core.typingWidget.SetNextSentence()
		core.statusWidget.Init(core.typingWidget.GetSentence())
		core.typingWidget.ClearInputBox()
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
			if core.statusWidget.Status.AmiWrong[i] == true {
				core.statusWidget.Status.WrongCount--
			}
			core.statusWidget.Status.AmiWrong[i] = false
		} else {
			colored += "[red]" + string(target[i])
			if core.statusWidget.Status.AmiWrong[i] == false {
				core.statusWidget.Status.AmiWrong[i] = true
				core.statusWidget.Status.WrongCount++
			}
		}
	}
	colored += "[-]"
	for i := len(curr); i < len(target); i++ {
		colored += string(target[i])
	}

	return
}
