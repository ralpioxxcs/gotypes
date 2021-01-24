package widget

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"strings"
)

const (
	corporate = `Leverage agile frameworks to provide a robust synopsis for high level overviews`
	wise      = `A bad workman always blames his tools`
)

// TypingBox is a box which display words be typed
// it include tview TextView , InputField struct
type TypingWidget struct {
	*tview.Flex
	Text     *tview.TextView
	Input    *tview.InputField
	words    []string
	sentence []string
	count    int
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

// GetSentence returns string of current typing senctence in box
func (t *TypingWidget) GetSentence() string {
	return t.sentence[t.count]
}

// GetWords returns slice of words
func (t *TypingWidget) GetWords() []string {
	return t.words
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
		AddItem(t.Text, 0, 10, false).
		AddItem(t.Input, 0, 1, true)

	t.sentence = append(t.sentence, corporate)

	t.SetTitle("TypingWidget")
	t.Text.SetText("\n\n" + t.sentence[t.count] + "\n\n")
	t.Text.SetTextAlign(tview.AlignCenter)

	t.words = strings.Split(t.sentence[t.count], " ")

	// configure function to typing box input field
	//t.input.SetChangedFunc(testos)

	return t
}
