package widget

import (
	_ "bufio"
	"encoding/json"
	_ "fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"io/ioutil"
	"os"
	_ "strings"
)

type Words struct {
	English     []string `json:"english"`
	Korean      []string `json:"korean"`
	English1000 []string `json:"english1000"`
}

// TypingBox is a box which display words be typed
// it include tview TextView , InputField struct
type TypingWidget struct {
	*tview.Flex
	Text     *tview.TextView
	Input    *tview.InputField
	Word     Words
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

func (t *TypingWidget) SetNextSentence() {
	t.count += 1
	t.Text.SetText("\n\n" + t.sentence[t.count] + "\n\n")
}

func (t *TypingWidget) ClearInputBox() {
	t.Input.SetText("")
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

	jsonFile, err := os.Open("data/test.json")
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &t.Word)

	//// read sentences from file
	//file, err := os.Open("data/wise-saying.txt")
	//if err != nil {
	//  panic(err)
	//}
	//defer file.Close()

	//var lines []string
	//scanner := bufio.NewScanner(file)
	//for scanner.Scan() {
	//  lines = append(lines, scanner.Text())
	//}
	//t.sentence = make([]string, len(lines), 100)
	//copy(t.sentence, lines)

	t.SetTitle("TypingWidget")

	var wordlines string
	const num = 20
	for i := 0; i < num; i++ {
		wordlines = wordlines + " " + t.Word.English[i]
	}
	//for _, v := range t.Word.English {
	//  wordlines = wordlines + "  " + v
	//}
	t.Text.SetText(wordlines)
	//t.Text.SetText("\n\n" + t.sentence[t.count] + "\n\n")
	t.Text.SetTextAlign(tview.AlignCenter)

	return t
}
