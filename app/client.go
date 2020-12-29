package app

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/jroimartin/gocui"
)

// Run app
func Run() {
	g, err := initailize()
	if err != nil {
		log.Fatalln("Failed to initailize a GUI:", err)
		return
	}
	defer g.Close()

	g.MainLoop()
}

// initailize cui interface
func initailize() (*gocui.Gui, error) {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		return g, err
	}

	g.Cursor = true
	g.SetManagerFunc(layout)

	if err = keybinding(g); err != nil {
		return g, err
	}
	return g, nil
}

func layout(g *gocui.Gui) error {

	/*
			    [ Layout ]

			-----------------
			 t    typing
			 h
			 e ______________
			 m
		 	 e 	 live wpm
			_________________
	*/

	maxX, maxY := g.Size()

	// pane 1 - theme
	theme, err := g.SetView("theme", -1, -1, 30, maxY)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	theme.Highlight = true
	theme.SelBgColor = gocui.ColorGreen
	theme.SelFgColor = gocui.ColorBlack
	fmt.Fprintln(theme, "8008")
	fmt.Fprintln(theme, "Dracula")

	// pane 2.1 - sentence viewport
	sviewport, err := g.SetView("sentence", 30, 0, maxX, maxY)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	sviewport.Title = "viewport"
	b, err := ioutil.ReadFile("../data/charlieandchcolatefactory.txt")
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(sviewport, "%s", b)

	// pane 2.2 - typing viewport
	_, h := sviewport.Size()
	tviewport, err := g.SetView("typing", 30, h/2, maxX, maxY)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	tviewport.Frame = true
	tviewport.Title = "types here"
	tviewport.FgColor = gocui.ColorYellow
	tviewport.Editable = true
	tviewport.Wrap = true

	if _, err = g.SetCurrentView("typing"); err != nil {
		return err
	}

	return nil
}

// keybinding configurate key bindings each pane
func keybinding(g *gocui.Gui) error {
	g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return gocui.ErrQuit
	})
	return nil
}
