package widget

import (
	"github.com/gdamore/tcell/v2"
	"strconv"
	"strings"
)

var themesString = []string{
	"Dark",
	"Dracula",
	"Dots",
	"Dualshot",
	"Oblivion",
	"Olive",
	"Bento",
	"Laser",
	"Hammerhead",
	"8008",
	"Nautilus",
}

type themes map[string]palette

func NewThemes() themes {
	return themes{
		"Dark":       Dark(),
		"Dracula":    Dracula(),
		"Dots":       Dots(),
		"Dualshot":   Dualshot(),
		"Oblivion":   Oblivion(),
		"Olive":      Olive(),
		"Bento":      Bento(),
		"Laser":      Laser(),
		"Hammerhead": Hammerhead(),
		"8008":       t8008(),
		"Nautilus":   Nautilus(),
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

func (p *palette) ToHexString() (title, background, foreground, border, extra string) {
	title = strconv.FormatUint(uint64(p.title), 16)
	background = strconv.FormatUint(uint64(p.background), 16)
	foreground = strconv.FormatUint(uint64(p.foreground), 16)
	border = strconv.FormatUint(uint64(p.border), 16)
	extra = strconv.FormatUint(uint64(p.extra), 16)

	foreground = strings.ReplaceAll(foreground, "300", "")
	extra = strings.ReplaceAll(extra, "300", "")
	return title, background, foreground, border, extra
}

func (p *palette) GetBg() tcell.Color {
	return p.background
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

// Olive returns gmk olive theme colour code
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

// Bento returns gmk bento theme colour code
func Bento() palette {
	return palette{
		name:       "Bento",
		title:      tcell.GetColor("#ff7a90"),
		background: tcell.GetColor("#2d394d"),
		foreground: tcell.GetColor("#fffaf8"),
		border:     tcell.GetColor("#4a768d"),
		extra:      tcell.GetColor("#ee2a3a"),
	}
}

// Laser returns gmk laser theme colour code
func Laser() palette {
	return palette{
		name:       "Laser",
		title:      tcell.GetColor("#009eaf"),
		background: tcell.GetColor("#221b44"),
		foreground: tcell.GetColor("#dbe7e8"),
		border:     tcell.GetColor("#b82356"),
		extra:      tcell.GetColor("#a8d400"),
	}
}

// Hammerhead returns gmk hammerhead theme colour code
func Hammerhead() palette {
	return palette{
		name:       "Hammerhead",
		title:      tcell.GetColor("#4fcdb9"),
		background: tcell.GetColor("#030613"),
		foreground: tcell.GetColor("#e2f1f5"),
		border:     tcell.GetColor("#1e283a"),
		extra:      tcell.GetColor("#e32b2b"),
	}
}

// t8008 returns gmk 8008 theme colour code
func t8008() palette {
	return palette{
		name:       "8008",
		title:      tcell.GetColor("#f44c7f"),
		background: tcell.GetColor("#333a45"),
		foreground: tcell.GetColor("#e9ecf0"),
		border:     tcell.GetColor("#939eae"),
		extra:      tcell.GetColor("#da3333"),
	}
}

// Nautilus returns gmk nautilus theme colour code
func Nautilus() palette {
	return palette{
		name:       "Nautilus",
		title:      tcell.GetColor("#ebb723"),
		background: tcell.GetColor("#132237"),
		foreground: tcell.GetColor("#1cbaac"),
		border:     tcell.GetColor("#0b4c6c"),
		extra:      tcell.GetColor("#da3333"),
	}
}
