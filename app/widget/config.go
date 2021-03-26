package widget

import (
	"github.com/rivo/tview"
)

type ConfigWidget struct {
	*tview.Flex
	LanguageList *tview.DropDown
	SoundList    *tview.DropDown
	language     int
	sound        bool
}

const (
	English = 0
	Korean
)

var languages = []string{
	"English",
	"Korean",
}

func (w *ConfigWidget) ApplyColor(p palette) {
	w.SetTitleColor(p.title)
	w.SetBorderColor(p.border)

	w.LanguageList.SetBackgroundColor(p.background)
	w.LanguageList.SetLabelColor(p.title)
	w.LanguageList.SetFieldTextColor(p.foreground)
	w.LanguageList.SetFieldBackgroundColor(p.border)
}

func (w *ConfigWidget) SetLanguage(lang int) {
	w.language = lang
}

func NewConfigWidget() *ConfigWidget {
	c := &ConfigWidget{
		Flex:         tview.NewFlex(),
		LanguageList: tview.NewDropDown(),
		SoundList:    tview.NewDropDown(),
		language:     English,
		sound:        false,
	}

	c.LanguageList.SetLabel("# Language :").
		SetOptions(languages, nil)
	//c.SoundList.SetLabel("sound").
	//  SetOptions([]string{"on", "off"}, nil)

	c.AddItem(c.LanguageList, 0, 1, false)
	//c.AddItem(c.LanguageList, 0, 1, false).
	//  AddItem(c.SoundList, 0, 1, true)

	c.SetBorder(true)
	c.SetTitle("Configuration")

	return c
}
