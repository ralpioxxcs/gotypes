package widget

import (
	//"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ConfigWidget struct {
	*tview.Flex
	LanguageList  *tview.DropDown
	SoundList     *tview.DropDown
	WordCountList *tview.DropDown

	optionLists map[int]*tview.DropDown
	currentOpt  int

	wordcount int
	language  int
	sound     bool
}

const (
	English = 0
	Korean
)

func (w *ConfigWidget) ApplyColor(p palette) {
	w.SetBackgroundColor(p.background)
	w.SetTitleColor(p.title)
	w.SetBorderColor(p.border)

	w.WordCountList.SetBackgroundColor(p.background)
	w.WordCountList.SetLabelColor(p.title)
	w.WordCountList.SetFieldTextColor(p.foreground)
	w.WordCountList.SetFieldBackgroundColor(p.border)
	w.WordCountList.SetPrefixTextColor(p.border)

	w.LanguageList.SetBackgroundColor(p.background)
	w.LanguageList.SetLabelColor(p.title)
	w.LanguageList.SetFieldTextColor(p.foreground)
	w.LanguageList.SetFieldBackgroundColor(p.border)
	w.LanguageList.SetPrefixTextColor(p.border)

	w.SoundList.SetBackgroundColor(p.background)
	w.SoundList.SetLabelColor(p.title)
	w.SoundList.SetFieldTextColor(p.foreground)
	w.SoundList.SetFieldBackgroundColor(p.border)
	w.SoundList.SetPrefixTextColor(p.border)
}

func (w *ConfigWidget) SetLanguage(lang int) {
	w.language = lang
}

func (w *ConfigWidget) GetNextOption() (opt *tview.DropDown) {
	w.currentOpt = (w.currentOpt + 1) % len(w.optionLists)
	return w.optionLists[w.currentOpt]
}

func NewConfigWidget() *ConfigWidget {
	c := &ConfigWidget{
		Flex:          tview.NewFlex(),
		WordCountList: tview.NewDropDown(),
		LanguageList:  tview.NewDropDown(),
		SoundList:     tview.NewDropDown(),
		currentOpt:    0,
		language:      English,
		sound:         false,
	}

	c.optionLists = map[int]*tview.DropDown{
		0: c.WordCountList,
		1: c.LanguageList,
		2: c.SoundList,
	}

	c.WordCountList.SetLabel("# Words : ").
		SetOptions([]string{"15", "30", "60", "120"}, nil)
	c.LanguageList.SetLabel("# Language :  ").
		SetOptions([]string{"English", "Korean"}, nil)
	c.SoundList.SetLabel("# Sound : ").
		SetOptions([]string{"off", "on"}, nil)

	c.WordCountList.SetCurrentOption(0)
	c.LanguageList.SetCurrentOption(0)
	c.SoundList.SetCurrentOption(0)

	c.AddItem(c.WordCountList, 0, 1, true).
		AddItem(c.LanguageList, 0, 1, true).
		AddItem(c.SoundList, 0, 1, true)

	c.SetBorder(true)
	c.SetTitle("Configuration")

	return c
}
