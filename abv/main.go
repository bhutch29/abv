package main

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"log"
)

var (
	g *gocui.Gui
)

const (
	logView      = "Log"
	input        = "Input"
	info         = "Info"
	popup        = "Popup"
	prompt       = "Prompt"
	promptSymbol = "PromptSymbol"
	errorView    = "Errors"
)

var keys = []key{
	{"", gocui.KeyCtrlC, quit, "C-c", "quit"},
	{input, gocui.KeyEnter, parseInput, "Enter", "confirm"},
	{popup, gocui.KeyArrowUp, popupScrollUp, "Up", "scrollUp"},
	{popup, gocui.KeyArrowDown, popupScrollDown, "Down", "scrollDown"},
	{popup, gocui.KeyEnter, popupSelectItem, "Enter", "Select"},
}

func main() {
	gui, err := gocui.NewGui(gocui.OutputNormal)
	g = gui
	if err != nil {
		log.Fatalln(err)
	}
	defer g.Close()

	g.SetManagerFunc(layout)
	g.Cursor = true

	if err := configureKeys(); err != nil {
		log.Fatalln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Fatalln(err)
	}
}

func logToMainView(s ...interface{}) error {
	v, err := g.View(logView)
	if err != nil {
		return err
	}
	fmt.Fprintln(v, s)
	return nil
}

func parseInput(g *gocui.Gui, v *gocui.View) error {
	vn, err := g.View(logView)
	if err != nil {
		displayError(err)
	}
	fmt.Fprintf(vn, "You typed: "+v.Buffer())
	updatePopup(v.Buffer())
	togglePopup()
	clearInput()
	return nil
}

func updatePopup(name string) {
	g.Update(func(g *gocui.Gui) error {
		v, err := g.View(popup)
		if err != nil {
			return err
		}

		drinks, err := SearchUntappdByName(name)
		if err != nil {
			logToMainView(err)
			displayError(err)
			return nil
		}
		hideError()
		v.Clear()

		for _, drink := range drinks {
			fmt.Fprintf(v, "%s: %s\n", drink.Brand, drink.Name)
		}

		return nil
	})
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
