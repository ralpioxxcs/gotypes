package app

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"log"
	"os"
)

const corporate = `Leverage agile frameworks to provide a robust synopsis for high level overviews. Iterative approaches to corporate strategy foster collaborative thinking to further the overall value proposition. Organically grow the holistic world view of disruptive innovation via workplace diversity and empowerment.
Bring to the table win-win survival strategies to ensure proactive domination. At the end of the day, going forward, a new normal that has evolved from generation X is on the runway heading towards a streamlined cloud solution. User generated content in real-time will have multiple touchpoints for offshoring.
Capitalize on low hanging fruit to identify a ballpark value added activity to beta test. Override the digital divide with additional clickthroughs from DevOps. Nanotechnology immersion along the information highway will close the loop on focusing solely on the bottom line.
[yellow]Press Enter, then Tab/Backtab for word selections`

var app *App
var Logger log.Logger

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

	return l
}

func (t *ThemeList) ApplyColor(p palette) {
	t.SetBackgroundColor(p.background)
	t.SetTitleColor(p.title)
	t.SetBorderColor(p.border)
	t.SetMainTextColor(p.foreground)
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
// it include tview TextView , InputField struct
type TypingWidget struct {
	*tview.Flex
	text   *tview.TextView
	input  *tview.InputField
	buffer string
}

func (t *TypingWidget) ApplyColor(p palette) {
	t.SetTitleColor(p.title)
	t.text.SetBackgroundColor(p.background)
	t.text.SetTextColor(p.foreground)
	t.text.SetBorderColor(p.border)
	t.input.SetBackgroundColor(p.background)
	t.input.SetFieldTextColor(p.foreground)
	t.input.SetFieldBackgroundColor(p.border)
	t.input.SetBorderColor(p.border)
}

func NewTypingWidget() *TypingWidget {
	t := &TypingWidget{
		Flex:   tview.NewFlex(),
		text:   tview.NewTextView(),
		input:  tview.NewInputField(),
		buffer: corporate,
	}

	t.text.SetBorder(true)

	t.input.
		SetPlaceholder("Type to start").
		SetLabelWidth(0).
		SetFieldWidth(0).
		SetPlaceholderTextColor(tcell.ColorBlack).
		SetFieldBackgroundColor(tcell.ColorGold).
		SetFieldTextColor(tcell.ColorBlack)
	t.input.SetBorder(true)

	t.SetDirection(tview.FlexRow).
		AddItem(t.text, 0, 10, false).
		AddItem(t.input, 0, 1, true)

	t.SetTitle("TypingWidget")
	t.text.SetText("\n\n" + t.buffer + "\n\n")
	t.text.SetTextAlign(tview.AlignCenter)

	t.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key != nil {
			app.SetFocus(t.input)
		}
		return event
	})

	return t
}

//--------------------------------------------------------------------

// StatsWidget is frame which display general typing information ( wpm, time ,,)
// it include tview Box struct
type StatsWidget struct {
	*tview.Box
	st Stats
}

// ApplyColor set color
func (t *StatsWidget) ApplyColor(p palette) {
	t.SetBackgroundColor(p.background)
	t.SetTitleColor(p.title)
	t.SetBorderColor(p.border)
}

func NewStatusWidget() *StatsWidget {
	d := &StatsWidget{
		Box: tview.NewBox(),
	}
	d.SetBorder(true)
	d.SetTitle("Result")
	return d
}

//--------------------------------------------------------------------

/* App is entire tui struct including tview flex struct
 * it consists of each widgets (sidebar, body, status)
 * - sidebar : it lists color themes
 * - body    : display words and current carrot interactively
 * - stats   : it shows current status such as wpm, time ..
 */
type App struct {
	*tview.Application
	flex    *tview.Flex
	sidebar *ThemeList
	body    *TypingWidget
	stats   *StatsWidget
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
func (a *App) menuAction(action MenuAction) {
	switch action {
	case MenuActionNone:
	case MenuActionImportTheme:
		a.sidebar.ApplyColor(a.sidebar.theme[a.sidebar.GetCurrentTheme()])
		a.body.ApplyColor(a.sidebar.theme[a.sidebar.GetCurrentTheme()])
		a.stats.ApplyColor(a.sidebar.theme[a.sidebar.GetCurrentTheme()])
	}
}

// NewApp returns initialized App struct
func NewApp() *App {
	a := &App{
		Application: tview.NewApplication(),
		flex:        tview.NewFlex(),
		sidebar:     NewThemeList(),
		body:        NewTypingWidget(),
		stats:       NewStatusWidget(),
	}

	// set function to sidebar
	a.sidebar.SetActionFunc(a.menuAction)

	// set body frame
	a.flex.AddItem(a.sidebar, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(a.body, 0, 8, true).
			AddItem(a.stats, 0, 2, false), 0, 9, true)
	a.menuAction(MenuActionImportTheme)

	a.SetRoot(a.flex, true)
	a.EnableMouse(true)

	// pass var's address to global
	app = a

	return a
}

func Run(a *App) {
	// logger
	f, err := os.OpenFile("log.txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		Logger.Fatal(err)
	}
	defer f.Close()
	Logger.SetOutput(f)
	Logger.Println("-----------------------------")

	// Run tui main
	if err := a.Run(); err != nil {
		panic(err)
	}
}
