package app

import (
	c "github.com/jroimartin/gocui"
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

	// 1. Viewport
	viewport, err := g.SetView("viewport", lw+1, 0, tw-1, th-ih-1)
	if err != nil && err != c.ErrUnknownView {
		return err
	}
	viewport.Title = "Viewport"
	viewport.BgColor = c.ColorGreen
	viewport.FgColor = c.ColorWhite
	viewport.Autoscroll = true

	/* paragraphs init
	 *
	 *
	 *
	 */

	// 2. Input
	input, err := g.SetView("input", lw+1, th-ih, tw-1, th-1)
	if err != nil && err != c.ErrUnknownView {
		return err
	}
	input.Title = "Input"
	input.BgColor = c.ColorBlack
	input.FgColor = c.ColorYellow
	input.Editable = true
	if err = input.SetCursor(0, 0); err != nil {
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
