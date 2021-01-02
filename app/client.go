package app

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"log"
	"os"
	"time"
)

const corporate = `Leverage agile frameworks to provide a robust synopsis for high level overviews. Iterative approaches to corporate strategy foster collaborative thinking to further the overall value proposition. Organically grow the holistic world view of disruptive innovation via workplace diversity and empowerment.
Bring to the table win-win survival strategies to ensure proactive domination. At the end of the day, going forward, a new normal that has evolved from generation X is on the runway heading towards a streamlined cloud solution. User generated content in real-time will have multiple touchpoints for offshoring.
Capitalize on low hanging fruit to identify a ballpark value added activity to beta test. Override the digital divide with additional clickthroughs from DevOps. Nanotechnology immersion along the information highway will close the loop on focusing solely on the bottom line.
[yellow]Press Enter, then Tab/Backtab for word selections`

var app tview.Application
var color palette

var logger log.Logger

//--------------------------------------------------------------------
type MenuAction int

const (
	MenuActionNone MenuAction = iota
	MenuActionImportTheme
)

// ThemeList is a box which display text theme list
type ThemeList struct {
	*tview.List
	theme    themes
	doneFunc func(action MenuAction)
}

func NewThemeList() *ThemeList {
	l := &ThemeList{
		List:  tview.NewList(),
		theme: NewThemes(),
	}
	l.ShowSecondaryText(false)
	l.SetBorder(true)
	l.SetTitle("Themes")

	for key, _ := range l.theme {
		l.AddItem(key, "", 0, l.doApply)
	}

	l.ApplyColor(l.theme["Dots"])

	return l
}

func (t *ThemeList) ApplyColor(p palette) {
	t.SetBackgroundColor(p.background)
	t.SetTitleColor(p.title)
	t.SetBorderColor(p.border)
	t.SetSelectedTextColor(p.extra)
}

func (t *ThemeList) GetCurrentTheme() string {
	name, _ := t.GetItemText(t.GetCurrentItem())
	return name
}

func (t *ThemeList) SetDoneFunc(doneFunc func()) {
}

func (t *ThemeList) SetActionFunc(doneFunc func(action MenuAction)) {
	t.doneFunc = doneFunc
}

func (t *ThemeList) SetVisible(visible bool) {
}

func (t *ThemeList) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		key := event.Key()
		if key == tcell.KeyEscape {
			if t.doneFunc != nil {
				t.doneFunc(MenuActionNone)
			}
		} else {
			t.List.InputHandler()(event, setFocus)
		}
	}
}

func (t *ThemeList) doApply() {
	if (t.doneFunc) != nil {
		t.doneFunc(MenuActionImportTheme)
	}
}

//--------------------------------------------------------------------

// TypingBox is a box which display words be typed
type TypingBox struct {
	*tview.TextView
	colors palette
	buffer string
}

func NewTypingBox() *TypingBox {
	t := &TypingBox{
		TextView: tview.NewTextView(),
		colors:   color,
		buffer:   corporate,
	}
	t.SetBorder(true)
	t.SetTitle("TypingBox")
	t.SetText("\n\n" + t.buffer + "\n\n")
	t.SetTextAlign(tview.AlignCenter)

	return t
}

//--------------------------------------------------------------------

// Dashboard is a frame which display general typing information ( wpm, time ,,)
type Dashboard struct {
	*tview.Box
	colors  palette
	wpm     int
	eplased time.Time
}

func NewDashboard() *Dashboard {
	d := &Dashboard{
		Box:    tview.NewBox(),
		colors: color,
		wpm:    0,
	}
	d.SetBorder(true)
	d.SetTitle("Result")
	return d
}

//--------------------------------------------------------------------
type App struct {
	app     *tview.Application
	flex    *tview.Flex
	sidebar *ThemeList
	body    *TypingBox
	result  *Dashboard
}

func (a *App) Draw(screen tcell.Screen) {
	a.flex.Draw(screen)
}

func (a *App) GetRect() (int, int, int, int) {
	return a.flex.GetRect()
}

//func (a *App) SetRect() (x, y, width, height int) {
//  a.flex.SetRect(x, y, width, height)
//}

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

func (a *App) menuAction(action MenuAction) {
	switch action {
	case MenuActionNone:
	case MenuActionImportTheme:
		a.sidebar.ApplyColor(a.sidebar.theme[a.sidebar.GetCurrentTheme()])
	}
}

//--------------------------------------------------------

func NewApp() *App {
	a := &App{
		app:     tview.NewApplication(),
		flex:    tview.NewFlex(),
		sidebar: NewThemeList(),
		body:    NewTypingBox(),
		result:  NewDashboard(),
	}

	a.sidebar.SetActionFunc(a.menuAction)

	a.flex.AddItem(a.sidebar, 0, 1, true).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(a.body, 0, 8, true).
			AddItem(a.result, 0, 2, false), 0, 9, false)

	a.app.SetRoot(a.flex, true)

	return a
}

func Run(a *App) {
	// logging
	f, err := os.OpenFile("filename", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		logger.Fatal(err)
	}
	defer f.Close()
	logger.SetOutput(f)
	logger.Println("-----------------------------")

	if err := a.app.Run(); err != nil {
		panic(err)
	}
}

func applyColor() {
}
