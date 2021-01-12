package app

import (
	"fmt"
)

func Test() {
	fmt.Println("wpm calcu")
}

// paintDiff returns an tview-painted string displaying the difference
func paintDiff(toColor string, differ string) (colorText string) {
	var h string

	for i := range differ {
		if i >= len(toColor) || differ[i] != toColor[i] {
			colorText += "[" + h + "red]"
		} else {
			colorText += "[" + h + "green]"
		}

		switch settings.I.ErrorDisplay {
		case settings.ErrorDisplayText:
			colorText += string(differ[i])
		case settings.ErrorDisplayTyped:
			if i < len(toColor) {
				colorText += string(toColor[i])
			}
		}
	}
	colorText += "[-:-:-]"

	if len(differ) < len(toColor) {
		colorText += toColor[len(differ):]
	}

	return
}
