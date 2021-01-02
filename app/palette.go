package app

import (
	"github.com/gdamore/tcell"
)

var themes = []string{
	"Dark",
	"Dracula",
	"Dots",
	"Dualshot",
	"Oblivion",
}

type Palette struct {
	name       string
	title      tcell.Color
	background tcell.Color
	foreground tcell.Color
	border     tcell.Color
	extra      tcell.Color
}

// NewPalette returns a new palette
func NewPalette() *Palette {
	palette := &Palette{
		name: "Dark",
	}
	palette.Dark()
	return palette
}

//
func (p *Palette) Dark() *Palette {
	p.name = "Dark"
	p.title = tcell.GetColor("#eee")
	p.background = tcell.GetColor("#111")
	p.foreground = tcell.GetColor("#eee")
	p.border = tcell.GetColor("444")
	p.extra = tcell.GetColor("da3333")
	return p
}

// Dracula returns dracula theme colour code
func (p *Palette) Dracula() *Palette {
	p.name = "Dracula"
	p.title = tcell.GetColor("#f2f2f2")
	p.background = tcell.GetColor("#282a36")
	p.foreground = tcell.GetColor("f2f2f2")
	p.border = tcell.GetColor("bd93f9")
	p.extra = tcell.GetColor("ff79c6")
	return p
}

// Dots returns dots theme colour code
func (p *Palette) Dots() *Palette {
	p.name = "Dots"
	p.title = tcell.GetColor("#fff")
	p.background = tcell.GetColor("#121520")
	p.foreground = tcell.GetColor("fff")
	p.border = tcell.GetColor("676e8a")
	p.extra = tcell.GetColor("791717")
	return p
}

// Dualshot returns dualshot theme colour code
func (p *Palette) Dualshot() *Palette {
	p.name = "Dualshot"
	p.title = tcell.GetColor("#737373")
	p.background = tcell.GetColor("#212222")
	p.foreground = tcell.GetColor("212222")
	p.border = tcell.GetColor("aaaaaa")
	p.extra = tcell.GetColor("c82931")
	return p
}

// Oblivion returns oblivion theme colour code
func (p *Palette) Oblivion() *Palette {
	p.name = "Oblivion"
	p.title = tcell.GetColor("#a5a096")
	p.background = tcell.GetColor("#313231")
	p.foreground = tcell.GetColor("f7f5f1")
	p.border = tcell.GetColor("5d6263")
	p.extra = tcell.GetColor("dd452e")
	return p
}

//// ModernDolchLight returns moderndolchlight theme colour code
//func (p *Palette) ModernDolchLight() *Palette {
//  p.title = tcell.GetColor("#a5a096")
//  p.background = tcell.GetColor("#313231")
//  p.foreground = tcell.GetColor("f7f5f1")
//  p.border = tcell.GetColor("5d6263")
//  p.extra = tcell.GetColor("dd452e")
//  return p
//}
