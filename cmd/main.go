package main

import (
	"github.com/ralpioxxcs/gotypes/app"
)

func main() {
	app.Run()
	return
}

//const (
//  lw = 20
//  ih = 3
//)

//var listItems = []string{
//  "Line 1",
//  "Line 2",
//  "Line 3",
//}

//func runGocui() {
//  g, err := c.NewGui(c.OutputNormal)
//  if err != nil {
//    log.Println("Failed to create a GUI:", err)
//    return
//  }
//  defer g.Close()

//  g.Cursor = true

//  g.SetManagerFunc(layout)

//  err = g.SetKeybinding("", c.KeyCtrlC, c.ModNone, quit)
//  if err != nil {
//    log.Println("Could not set key bindings:", err)
//    return
//  }

//  tw, th := g.Size()

//  // List
//  lv, err := g.SetView("list", 0, 0, lw, th-1)
//  if err != nil && err != c.ErrUnknownView {
//    log.Println("Failed to create main view:", err)
//    return
//  }
//  lv.Title = "List"
//  lv.FgColor = c.ColorCyan

//  // Output
//  ov, err := g.SetView("output", lw+1, 0, tw-1, th-ih-1)
//  if err != nil && err != c.ErrUnknownView {
//    log.Println("Failed to create output view:", err)
//    return
//  }
//  ov.Title = "Output"
//  ov.FgColor = c.ColorGreen
//  ov.Autoscroll = true
//  _, err = fmt.Fprintln(ov, "Press Ctrl+C to quit")
//  if err != nil {
//    log.Println("Failed to print into output view:", err)
//  }

//  // Input
//  iv, err := g.SetView("input", lw+1, th-ih, tw-1, th-1)
//  if err != nil && err != c.ErrUnknownView {
//    log.Println("Failed to create input view:", err)
//    return
//  }
//  iv.Title = "Input"
//  iv.FgColor = c.ColorYellow
//  iv.Editable = true
//  err = iv.SetCursor(0, 0)
//  if err != nil {
//    log.Println("Failed to set cursor:", err)
//    return
//  }

//  err = g.SetKeybinding("input", c.KeyEnter, c.ModNone, func(g *c.Gui, iv *c.View) error {

//    iv.Rewind()

//    ov, e := g.View("output")
//    if e != nil {
//      log.Println("Can't get output view:", e)
//      return e
//    }

//    _, e = fmt.Fprint(ov, iv.Buffer())
//    if e != nil {
//      log.Println("Can't print to output view:", e)
//    }
//    iv.Clear()

//    e = iv.SetCursor(0, 0)
//    if e != nil {
//      log.Println("Failed to set cursor:", e)
//    }
//    return e
//  })
//  if err != nil {
//    log.Println("Can't bind the enter key:", err)
//  }

//  for _, s := range listItems {
//    _, err = fmt.Fprintln(lv, s)
//    if err != nil {
//      log.Println("Error writing to the list view:", err)
//      return
//    }
//  }

//  _, err = g.SetCurrentView("input")
//  if err != nil {
//    log.Println("Can't set focus to input view:", err)
//  }

//  err = g.MainLoop()
//  log.Println("Main loop has finished:", err)
//}

//func layout(g *c.Gui) error {
//  tw, th := g.Size()

//  _, err := g.SetView("list", 0, 0, lw, th-1)
//  if err != nil {
//    return err
//  }
//  _, err = g.SetView("output", lw+1, 0, tw-1, th-ih-1)
//  if err != nil {
//    return err
//  }
//  _, err = g.SetView("input", lw+1, th-ih, tw-1, th-1)
//  if err != nil {
//    return err
//  }
//  return nil
//}

//func quit(g *c.Gui, v *c.View) error {
//  return c.ErrQuit
//}
