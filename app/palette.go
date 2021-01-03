package app

import (
	"github.com/gdamore/tcell"
)

var themesString = []string{
	"Dark",
	"Dracula",
	"Dots",
	"Dualshot",
	"Oblivion",
	"Olive",
}

type themes map[string]palette

func NewThemes() themes {
	return themes{
		"Dark":     Dark(),
		"Dracula":  Dracula(),
		"Dots":     Dots(),
		"Dualshot": Dualshot(),
		"Oblivion": Oblivion(),
		"Olive":    Olive(),
	}
}

type palette struct {
	name       string
	title      tcell.Color
	background tcell.Color
	foreground tcell.Color
	border     tcell.Color
	extra      tcell.Color
}

// Dark returns default colour code
func Dark() palette {
	return palette{
		name:       "Dark",
		title:      tcell.GetColor("#eeeeee"),
		background: tcell.GetColor("#111111"),
		foreground: tcell.GetColor("#eeeeee"),
		border:     tcell.GetColor("#444444"),
		extra:      tcell.GetColor("#da3333"),
	}
}

// Dracula returns dracula theme colour code
func Dracula() palette {
	return palette{
		name:       "Dracula",
		title:      tcell.GetColor("#f2f2f2"),
		background: tcell.GetColor("#282a36"),
		foreground: tcell.GetColor("#f2f2f2"),
		border:     tcell.GetColor("#bd93f9"),
		extra:      tcell.GetColor("#ff79c6"),
	}
}

// Dots returns dots theme colour code
func Dots() palette {
	return palette{
		name:       "Dots",
		title:      tcell.GetColor("#fff"),
		background: tcell.GetColor("#121520"),
		foreground: tcell.GetColor("#fff"),
		border:     tcell.GetColor("#676e8a"),
		extra:      tcell.GetColor("#791717"),
	}
}

// Dualshot returns dualshot theme colour code
func Dualshot() palette {
	return palette{
		name:       "Dualshot",
		title:      tcell.GetColor("#212222"),
		background: tcell.GetColor("#737373"),
		foreground: tcell.GetColor("#212222"),
		border:     tcell.GetColor("#aaaaaa"),
		extra:      tcell.GetColor("#c82931"),
	}
}

// Oblivion returns oblivion theme colour code
func Oblivion() palette {
	return palette{
		name:       "Oblivion",
		title:      tcell.GetColor("#a5a096"),
		background: tcell.GetColor("#313231"),
		foreground: tcell.GetColor("#f7f5f1"),
		border:     tcell.GetColor("#5d6263"),
		extra:      tcell.GetColor("#dd452e"),
	}
}

// Oblivion returns oblivion theme colour code
func Olive() palette {
	return palette{
		name:       "Olive",
		title:      tcell.GetColor("#92946f"),
		background: tcell.GetColor("#e9e5cc"),
		foreground: tcell.GetColor("#373731"),
		border:     tcell.GetColor("#b7b39e"),
		extra:      tcell.GetColor("#cf2f2f"),
	}
}

//// ModernDolchLight returns moderndolchlight theme colour code
//func ModernDolchLight
//  title = tcell.GetColor("#a5a096")
//  background = tcell.GetColor("#313231")
//  foreground = tcell.GetColor("f7f5f1")
//  border = tcell.GetColor("5d6263")
//  extra = tcell.GetColor("dd452e")
//  return p
//}
