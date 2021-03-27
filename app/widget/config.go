package widget

import (
	"github.com/rivo/tview"
)

type ConfigWidget struct {
	*tview.Flex
	LanguageList  *tview.DropDown
	SoundList     *tview.DropDown
	WordCountList *tview.DropDown
	wordcount     int
	language      int
	sound         bool
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

	w.WordCountList.SetBackgroundColor(p.background)
	w.WordCountList.SetLabelColor(p.title)
	w.WordCountList.SetFieldTextColor(p.foreground)
	w.WordCountList.SetFieldBackgroundColor(p.border)

	w.LanguageList.SetBackgroundColor(p.background)
	w.LanguageList.SetLabelColor(p.title)
	w.LanguageList.SetFieldTextColor(p.foreground)
	w.LanguageList.SetFieldBackgroundColor(p.border)

	w.SoundList.SetBackgroundColor(p.background)
	w.SoundList.SetLabelColor(p.title)
	w.SoundList.SetFieldTextColor(p.foreground)
	w.SoundList.SetFieldBackgroundColor(p.border)
}

func (w *ConfigWidget) SetLanguage(lang int) {
	w.language = lang
}

func NewConfigWidget() *ConfigWidget {
	c := &ConfigWidget{
		Flex:          tview.NewFlex(),
		WordCountList: tview.NewDropDown(),
		LanguageList:  tview.NewDropDown(),
		SoundList:     tview.NewDropDown(),
		language:      English,
		sound:         false,
	}

	c.WordCountList.SetLabel("# Words : ").
		SetOptions([]string{"15", "30", "60", "120"}, nil)
	c.LanguageList.SetLabel("# Language :").
		SetOptions(languages, nil)
	c.SoundList.SetLabel("# Sound").
		SetOptions([]string{"on", "off"}, nil)

	c.AddItem(c.WordCountList, 0, 1, false)
	c.AddItem(c.LanguageList, 0, 1, false)
	c.AddItem(c.SoundList, 0, 1, true)

	c.SetBorder(true)
	c.SetTitle("Configuration")

	return c
}
