package app

import (
	"fmt"
	c "github.com/jroimartin/gocui"
	"io/ioutil"
	"log"
)

const (
	lw = 20
	ih = 3
)

func Run() {
	g, err := initailize()
	if err != nil {
		log.Fatalln("Failed to initailize a GUI:", err)
		return
	}
	defer g.Close()

	g.MainLoop()
}

func initailize() (*c.Gui, error) {
	g, err := c.NewGui(c.OutputNormal)
	if err != nil {
		return g, err
	}

	g.Cursor = true

	g.SetManagerFunc(layout)

	if err = initKeybinding(g); err != nil {
		return g, err
	}
	return g, nil
}

func layout(g *c.Gui) error {
	tw, th := g.Size()

	maxX, maxY := tw-1, th-1

	// Typing window
	viewport, err := g.SetView("words", 0, 0, maxX, maxY/3)
	if err != nil && err != c.ErrUnknownView {
		return err
	}
	viewport.Title = "Typing Window"
	//viewport.BgColor = c.ColorGreen
	viewport.FgColor = c.ColorWhite
	viewport.Autoscroll = true

	// paragraphs init
	b, err := ioutil.ReadFile("../data/charlieandchcolatefactory.txt")
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(viewport, "%s", b)

	// 2. Input
	_, h := viewport.Size()
	input, err := g.SetView("input", 0, h/2, tw-1, h+1)
	if err != nil && err != c.ErrUnknownView {
		return err
	}
	input.Frame = true
	input.Title = "types here"
	//input.BgColor = c.ColorBlack
	input.FgColor = c.ColorYellow
	input.Editable = true
	input.Wrap = true
	if _, err = g.SetCurrentView("input"); err != nil {
		return err
	}

	return nil
}

func initKeybinding(g *c.Gui) error {
	g.SetKeybinding("", c.KeyCtrlC, c.ModNone, func(g *c.Gui, v *c.View) error {
		return c.ErrQuit
	})
	return nil
}
