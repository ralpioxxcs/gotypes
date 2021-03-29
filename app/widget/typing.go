package widget

import (
	"encoding/json"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"io/ioutil"
	"math/rand"
	"os"
	"time"
)

// languages contains words each languages
type languages struct {
	English     []string `json:"english"`
	Korean      []string `json:"korean"`
	English1000 []string `json:"english1000"`
}

// CopyTo copies slices element (deep copy)
func (w *languages) CopyTo() (dst languages) {
	dst.English = make([]string, len(w.English))
	dst.English1000 = make([]string, len(w.English1000))
	dst.Korean = make([]string, len(w.Korean))
	copy(dst.English, w.English)
	copy(dst.English1000, w.English1000)
	copy(dst.Korean, w.Korean)

	return dst
}

// TypingBox is a box which display words be typed
// It include struct of tview.TextView , tview.InputField
type TypingWidget struct {
	*tview.Flex
	Text         *tview.TextView
	Input        *tview.InputField
	Words        languages
	DisplayWords languages
	count        int
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
//
func (w *TypingWidget) Update(colored string, index int) {
	w.DisplayWords.English[index] = colored

	var wordlines string
	for i := 0; i < w.count; i++ {
		wordlines = wordlines + " " + w.DisplayWords.English[i]
	}

	w.Text.SetText("\n\n\n\n\n" + wordlines)
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
		count: 20,
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
		AddItem(w.Text, 20, 0, false).
		AddItem(w.Input, 3, 0, true)

	// load & display words
	jsonFile, err := os.Open("data/test.json")
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var allWords languages
	json.Unmarshal(byteValue, &allWords)

	// shuffle words
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(allWords.English), func(i, j int) {
		allWords.English[i], allWords.English[j] = allWords.English[j], allWords.English[i]
	})
	rand.Shuffle(len(allWords.English1000), func(i, j int) {
		allWords.English1000[i], allWords.English1000[j] = allWords.English1000[j], allWords.English1000[i]
	})
	rand.Shuffle(len(allWords.Korean), func(i, j int) {
		allWords.Korean[i], allWords.Korean[j] = allWords.Korean[j], allWords.Korean[i]
	})
	w.Words.English = make([]string, w.count)
	w.Words.English1000 = make([]string, w.count)
	w.Words.Korean = make([]string, w.count)
	copy(w.Words.English, allWords.English)
	copy(w.Words.English1000, allWords.English1000)
	copy(w.Words.Korean, allWords.Korean)

	w.DisplayWords = w.Words.CopyTo()

	var wordlines string
	for i := 0; i < w.count; i++ {
		wordlines = wordlines + " " + w.DisplayWords.English[i]
	}
	w.Text.SetText("\n\n\n\n\n" + wordlines)
	w.Text.SetTextAlign(tview.AlignCenter)

	return w
}
