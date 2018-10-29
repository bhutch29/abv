package main

import (
	"fmt"
	"log"

	"github.com/jroimartin/gocui"
)

var g gocui.Gui

var keys = []key{
	key{"", gocui.KeyCtrlC, quit, "C-c", "quit"},
	key{"", gocui.KeyEnter, parseInput, "Enter", "confirm"},
	key{"", gocui.KeyCtrl2, togglePopup, "C-2", "temporary"},
}

func main() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Fatalln(err)
	}
	defer g.Close()

	g.SetManagerFunc(layout)
	g.Cursor = true

	if err := configureKeys(g); err != nil {
		log.Fatalln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Fatalln(err)
	}
}

func logToMain(s interface{}) error {
	v, err := g.View("Main")
	if err != nil {
		return err
	}
	fmt.Fprintln(v, s)
	return nil
}

func parseInput(g *gocui.Gui, v *gocui.View) error {
	main, err := g.View("Main")
	if err != nil {
		return err
	}
	fmt.Fprintf(main, "You typed: "+v.Buffer())
	togglePopup(g, v)
	updatePopup(g, v.Buffer())
	clearInput(g)
	return nil
}

func updatePopup(g *gocui.Gui, name string) {
	g.Update(func(g *gocui.Gui) error {
		v, err := g.View("Popup")
		if err != nil {
			return err
		}

		drinks, err := SearchUntappdByName(name)
		if err != nil {
			displayError(g, err)
			return nil
		}
		hideError(g)

		v.Clear()

		for _, drink := range drinks {
			fmt.Fprintf(v, "%s: %s", drink.Brand, drink.Name)
		}

		return nil
	})
}

func getViewLine(g *gocui.Gui, v *gocui.View) (string, error) {
	var l string
	var err error

	_, cy := v.Cursor()
	if l, err = v.Line(cy); err != nil {
		l = ""
	}

	return l, err
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
