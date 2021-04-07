package app

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/ralpioxxcs/gotypes/app/widget"
	"github.com/rivo/tview"
)

/*
 * App is entire tui struct including tview flex struct
 * it consists of each widgets (sidebar, typing, status, config)
 * - sidebar : it lists color themes
 * - typing  : display words and current carrot interactively
 * - status  : it shows current status such as wpm, time ..
 * - config  : configurate several options such as word count, languages ..
 */
type App struct {
	*tview.Application
	flex          *tview.Flex
	blank         *tview.Box
	sidebarWidget *widget.ThemeList
	typingWidget  *widget.TypingWidget
	statusWidget  *widget.StatusWidget
	configWidget  *widget.ConfigWidget
	typingStarted bool
	quitLoop      chan bool
	Loopfinished  chan bool
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
		a.blank.SetBackgroundColor(bg.GetBg())
		a.sidebarWidget.ApplyColor(a.sidebarWidget.Theme[a.sidebarWidget.GetCurrentTheme()])
		a.typingWidget.ApplyColor(a.sidebarWidget.Theme[a.sidebarWidget.GetCurrentTheme()])
		a.statusWidget.ApplyColor(a.sidebarWidget.Theme[a.sidebarWidget.GetCurrentTheme()])
		a.configWidget.ApplyColor(a.sidebarWidget.Theme[a.sidebarWidget.GetCurrentTheme()])
	}
}

// Reset state
func (a *App) Reset(num int) {
	if a.typingStarted == true {
		a.quitLoop <- true
		<-a.Loopfinished

		a.typingStarted = false
		a.typingWidget.Input.SetChangedFunc(func(text string) {})
		//a.typingWidget.ClearInputBox()

		a.statusWidget.Reset()
		a.typingWidget.Reset()

		a.typingWidget.UpdateWords(num)
		a.typingWidget.Input.SetChangedFunc(startTyping)
	}
}

// NewApp returns initialized App struct
func NewApp() *App {
	a := &App{
		Application:   tview.NewApplication(),
		flex:          tview.NewFlex(),
		blank:         tview.NewBox(),
		sidebarWidget: widget.NewThemeList(),
		typingWidget:  widget.NewTypingWidget(),
		statusWidget:  widget.NewStatusWidget(),
		configWidget:  widget.NewConfigWidget(),
		typingStarted: false,
		quitLoop:      make(chan bool),
		Loopfinished:  make(chan bool),
	}

	// set function to side-bar widget
	a.sidebarWidget.SetActionFunc(a.menuAction)
	// config callback functions of typing widget
	// -> set focus to typing widget first time
	a.typingWidget.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		a.SetFocus(a.typingWidget.Input)
		return event
	})
	// set callback functions of config widget
	a.configWidget.WordCountList.SetSelectedFunc(func(text string, index int) {
		num, _ := strconv.Atoi(text)
		a.Reset(num)
	})
	a.configWidget.LanguageList.SetSelectedFunc(func(text string, index int) {
	})
	a.configWidget.SoundList.SetSelectedFunc(func(text string, index int) {
	})
	a.configWidget.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			a.SetFocus(a.configWidget.GetNextOption())
		}
		return event
	})
	// set callback functions of typing widget
	// -> realtime typing
	a.typingWidget.Input.SetChangedFunc(startTyping)
	// -> SetDoneFunc sets a handler which is called when the user is done entering text.
	a.typingWidget.Input.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyTab {
			_, text := a.configWidget.WordCountList.GetCurrentOption()
			num, _ := strconv.Atoi(text)
			a.Reset(num)
		}
	})

	// set app layout
	a.flex.AddItem(a.sidebarWidget, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(a.typingWidget, 0, 5, true).
			AddItem(a.blank, 0, 5, false).
			AddItem(a.statusWidget, 0, 2, false).
			AddItem(a.configWidget, 0, 2, false), 0, 10, true)
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

	Core = a

	return a
}

// App instance handler
var Core *App
var popup *tview.Modal

var cntt int

// startTyping process typing functions
// * text : current text
func startTyping(text string) {
	/*
	* store start time & elapsed
	* compare current words with indicating words
	 */

	if Core.typingStarted == false {
		Core.typingStarted = true
		Core.statusWidget.Init(Core.typingWidget.Words.English)
		go func() {
			timeout := time.After(60 * time.Second)
			// set AFK timeout (60 seconds) & update status widget each 50 miliseconds tick
			ticker := time.NewTicker(time.Millisecond * 50)
			for range ticker.C {
				select {
				case <-timeout:
					return
				case <-Core.quitLoop:
					Core.Loopfinished <- true
					return
				default:
					// Update status widget
					Core.QueueUpdateDraw(func() {
						Core.statusWidget.Wpm.
							SetText(fmt.Sprintf("Wpm : %.0f", Core.statusWidget.GetGrossWpm()))
						Core.statusWidget.Accuracy.
							SetText(fmt.Sprintf("Accuracy : %.0f", Core.statusWidget.GetAccuracy()))
						Core.statusWidget.Timer.
							SetText(fmt.Sprintf("Time : %.02f sec", Core.statusWidget.GetElapsed()))
						Core.statusWidget.Count.
							SetText(fmt.Sprintf("Count : %d", Core.statusWidget.GetCount()))
					})
				}
			}
		}()
	}

	// check to pass next word
	if Core.typingStarted == true {
		if (len(Core.statusWidget.Status.GetCurrentWord().Text)) < len(text) {
			runes := []rune(text)
			if runes[len(text)-1] == ' ' { // check whitespace
				Core.statusWidget.Status.AddCount()
				Core.typingWidget.ClearInputBox()
			}
			return
			//Core.SetRoot(popup, false).SetFocus(popup).Run()
		}

		// compare typingWidget word with target word & coloring , underlining
		palette := Core.sidebarWidget.Theme[Core.sidebarWidget.GetCurrentTheme()]
		_, _, fg, _, err := palette.ToHexString()

		Core.typingWidget.Update(
			diff(text, Core.statusWidget.Status.GetCurrentWord(), fg, err), Core.statusWidget.GetCount()-1)
	}
}

// diff returns colored string of current word
func diff(curr string, target widget.Word, textcolor string, textcolor_error string) (colored string) {
	for i := range curr {
		if curr[i] == target.Text[i] {
			//colored += "[green]" + string(curr[i])
			colored += "[#" + textcolor + "]" + string(curr[i])
			if target.Iswrong[i] == true {
				Core.statusWidget.Status.WrongEntries--
			}
			target.Iswrong[i] = false
		} else {
			//colored += "[red]" + string(target.Text[i])
			colored += "[#" + textcolor_error + "]" + string(target.Text[i])
			if target.Iswrong[i] == false {
				target.Iswrong[i] = true
				Core.statusWidget.Status.WrongEntries++
			}
		}
	}
	colored += "[-]"
	for i := len(curr); i < len(target.Text); i++ {
		colored += string(target.Text[i])
	}
	return colored
}
