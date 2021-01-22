package widget

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type MenuAction int

const (
	MenuActionNone MenuAction = iota
	MenuActionImportTheme
)

// ThemeList is a box which display text theme list
type ThemeList struct {
	*tview.List
	Theme    themes
	doneFunc func(action MenuAction)
}

func NewThemeList() *ThemeList {
	l := &ThemeList{
		List:  tview.NewList(),
		Theme: NewThemes(),
	}
	l.ShowSecondaryText(false)
	l.SetBorder(true)
	l.SetTitle("Themes")

	for key, _ := range l.Theme {
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
