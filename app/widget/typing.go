package widget

import (
	"encoding/json"
	_ "fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"io/ioutil"
	"os"
	_ "strings"
)

// word contains words each languages
type word struct {
	English     []string `json:"english"`
	Korean      []string `json:"korean"`
	English1000 []string `json:"english1000"`
}

// TypingBox is a box which display words be typed
// It include struct of tview.TextView , tview.InputField
type TypingWidget struct {
	*tview.Flex
	Text  *tview.TextView
	Input *tview.InputField
	Words word
	count int
}

// ApplyColor apply current theme color on widget
func (w *TypingWidget) ApplyColor(p palette) {
	w.SetTitleColor(p.title)

	w.Text.SetBackgroundColor(p.background)
	w.Text.SetTextColor(p.border)
	w.Text.SetBorderColor(p.border)

	w.Input.SetBackgroundColor(p.background)
	w.Input.SetFieldTextColor(p.foreground)
	w.Input.SetFieldBackgroundColor(p.border)
	w.Input.SetBorderColor(p.border)
}

// Update updates word list whether it is correct or not
func (w *TypingWidget) Update(colored string, index int) {
	var wordlines string

	wordlines += colored
	const num = 20
	for i := index + 1; i < num; i++ {
		wordlines = wordlines + " " + w.Words.English[i]
	}
	w.Text.SetText(wordlines)
	w.Text.SetTextAlign(tview.AlignCenter)
}

func (w *TypingWidget) ClearInputBox() {
	w.Input.SetText("")
}

func NewTypingWidget() *TypingWidget {
	w := &TypingWidget{
		Flex:  tview.NewFlex(),
		Text:  tview.NewTextView(),
		Input: tview.NewInputField(),
		count: 0,
	}

	w.Text.SetBorder(true)
	w.Text.SetDynamicColors(true)

	w.Input.
		SetPlaceholder("Type to start").
		SetLabelWidth(0).
		SetFieldWidth(0).
		SetPlaceholderTextColor(tcell.ColorBlack).
		SetFieldBackgroundColor(tcell.ColorGold).
		SetFieldTextColor(tcell.ColorBlack)
	w.Input.SetBorder(true)

	w.SetDirection(tview.FlexRow).
		AddItem(w.Text, 10, 0, false).
		AddItem(w.Input, 3, 0, true)

	// load & display words
	jsonFile, err := os.Open("data/test.json")
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &w.Words)

	var wordlines string
	const num = 20
	for i := 0; i < num; i++ {
		wordlines = wordlines + " " + w.Words.English[i]
	}
	w.Text.SetText(wordlines)
	w.Text.SetTextAlign(tview.AlignCenter)

	return w
}
