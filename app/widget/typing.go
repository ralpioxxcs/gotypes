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

type word struct {
	English     []string `json:"english"`
	Korean      []string `json:"korean"`
	English1000 []string `json:"english1000"`
}

// TypingBox is a box which display words be typed
// it include tview TextView , InputField struct
type TypingWidget struct {
	*tview.Flex
	Text  *tview.TextView
	Input *tview.InputField
	Words word
	count int
}

func (t *TypingWidget) ApplyColor(p palette) {
	t.SetTitleColor(p.title)

	t.Text.SetBackgroundColor(p.background)
	t.Text.SetTextColor(p.foreground)
	t.Text.SetBorderColor(p.border)

	t.Input.SetBackgroundColor(p.background)
	t.Input.SetFieldTextColor(p.foreground)
	t.Input.SetFieldBackgroundColor(p.border)
	t.Input.SetBorderColor(p.border)
}

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

func (t *TypingWidget) ClearInputBox() {
	t.Input.SetText("")
}

func NewTypingWidget() *TypingWidget {
	t := &TypingWidget{
		Flex:  tview.NewFlex(),
		Text:  tview.NewTextView(),
		Input: tview.NewInputField(),
		count: 0,
	}

	t.Text.SetBorder(true)
	t.Text.SetDynamicColors(true)

	t.Input.
		SetPlaceholder("Type to start").
		SetLabelWidth(0).
		SetFieldWidth(0).
		SetPlaceholderTextColor(tcell.ColorBlack).
		SetFieldBackgroundColor(tcell.ColorGold).
		SetFieldTextColor(tcell.ColorBlack)
	t.Input.SetBorder(true)

	t.SetDirection(tview.FlexRow).
		AddItem(t.Text, 10, 0, false).
		AddItem(t.Input, 3, 0, true)

	// load & display words
	jsonFile, err := os.Open("data/test.json")
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &t.Words)

	var wordlines string
	const num = 20
	for i := 0; i < num; i++ {
		wordlines = wordlines + " " + t.Words.English[i]
	}
	t.Text.SetText(wordlines)
	t.Text.SetTextAlign(tview.AlignCenter)

	return t
}
