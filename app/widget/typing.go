package widget

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

const corporate = `Leverage agile frameworks to provide a robust synopsis for high level overviews. Iterative approaches to corporate strategy foster collaborative thinking to further the overall value proposition. Organically grow the holistic world view of disruptive innovation via workplace diversity and empowerment.
Bring to the table win-win survival strategies to ensure proactive domination. At the end of the day, going forward, a new normal that has evolved from generation X is on the runway heading towards a streamlined cloud solution. User generated content in real-time will have multiple touchpoints for offshoring.
Capitalize on low hanging fruit to identify a ballpark value added activity to beta test. Override the digital divide with additional clickthroughs from DevOps. Nanotechnology immersion along the information highway will close the loop on focusing solely on the bottom line.`

// TypingBox is a box which display words be typed
// it include tview TextView , InputField struct
type TypingWidget struct {
	*tview.Flex
	Text     *tview.TextView
	Input    *tview.InputField
	sentence string
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

func (t *TypingWidget) GetSentence() string {
	return t.sentence
}

func NewTypingWidget() *TypingWidget {
	t := &TypingWidget{
		Flex:     tview.NewFlex(),
		Text:     tview.NewTextView(),
		Input:    tview.NewInputField(),
		sentence: corporate,
	}

	t.Text.SetBorder(true)

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

	t.SetTitle("TypingWidget")
	t.Text.SetText("\n\n" + t.sentence + "\n\n")
	t.Text.SetTextAlign(tview.AlignCenter)

	// configure function to typing box input field
	//t.input.SetChangedFunc(testos)

	return t
}
