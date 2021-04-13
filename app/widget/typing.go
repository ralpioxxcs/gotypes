package widget

import (
	"encoding/json"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
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
	CurrentIndex int
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

func (w *TypingWidget) Reset() {
	w.Words.English = nil
	w.Words.English1000 = nil
	w.Words.Korean = nil
	w.DisplayWords.English = nil
	w.DisplayWords.English1000 = nil
	w.DisplayWords.Korean = nil
	w.CurrentIndex = 1
	w.count = 0
}

func (w *TypingWidget) UpdateWords(number int) {
	w.ClearInputBox()
	w.count = number

	// load & display words
	jsonFile, err := os.Open("data/words.json")
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
		allWords.English[i], allWords.English[j] =
			allWords.English[j], allWords.English[i]
	})
	rand.Shuffle(len(allWords.English1000), func(i, j int) {
		allWords.English1000[i], allWords.English1000[j] =
			allWords.English1000[j], allWords.English1000[i]
	})
	rand.Shuffle(len(allWords.Korean), func(i, j int) {
		allWords.Korean[i], allWords.Korean[j] =
			allWords.Korean[j], allWords.Korean[i]
	})
	w.Words.English = make([]string, number)
	w.Words.English1000 = make([]string, number)
	w.Words.Korean = make([]string, number)
	copy(w.Words.English, allWords.English)
	copy(w.Words.English1000, allWords.English1000)
	copy(w.Words.Korean, allWords.Korean)

	w.DisplayWords = w.Words.CopyTo()

	var wordlines string
	for i := 0; i < w.count; i++ {
		word := fmt.Sprintf(`["%d"]%s[""]`, i, w.DisplayWords.English[i])
		if i != 0 {
			wordlines = wordlines + "\n" + word
		} else {
			wordlines = word
		}
	}
	w.Text.SetText(wordlines)
	w.Text.SetTextAlign(tview.AlignCenter)
}

// Update updates word list whether it is correct or not
func (w *TypingWidget) Update(colored string, index int) {
	w.DisplayWords.English[index] = colored

	var wordlines string
	for i := 0; i < w.count; i++ {
		word := fmt.Sprintf(`["%d"]%s[""]`, i, w.DisplayWords.English[i])
		if i != 0 {
			wordlines = wordlines + "\n" + word
		} else {
			wordlines = word
		}
	}
	w.Text.SetText(wordlines)
	w.Text.SetTextAlign(tview.AlignCenter)
}

func (w *TypingWidget) ProcessNextWord() {
	w.Text.Highlight(strconv.Itoa(w.CurrentIndex)).ScrollToHighlight()
}

func (w *TypingWidget) ClearInputBox() {
	w.Input.SetText("")
}

func NewTypingWidget() *TypingWidget {
	w := &TypingWidget{
		Flex:         tview.NewFlex(),
		Text:         tview.NewTextView(),
		Input:        tview.NewInputField(),
		CurrentIndex: 1,
		count:        60,
	}

	w.Text.SetBorder(true)
	w.Text.SetRegions(true)
	w.Text.SetScrollable(true)
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
		AddItem(w.Text, 0, 6, false).
		AddItem(w.Input, 0, 1, true)

	w.UpdateWords(w.count)

	return w
}
