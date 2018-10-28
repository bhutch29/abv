package main

import (
	"github.com/jroimartin/gocui"
	"log"
	"fmt"
)

func main() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Fatalln(err)
	}
	defer g.Close()

	g.SetManagerFunc(layout)
	g.Cursor = true

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Fatalln(err)
	}

	if err := g.SetKeybinding("", gocui.KeyEnter, gocui.ModNone, parseInput); err != nil {
		log.Fatalln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Fatalln(err)
	}
}

func parseInput(g *gocui.Gui, v *gocui.View) error {
	main, err := g.View("Main")
	if err != nil {
		return err
	}
	fmt.Fprintf(main, v.Buffer())
	printPrompt(g)
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
