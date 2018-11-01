package main

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/sirupsen/logrus"
	"os"
)

var (
	g       *gocui.Gui
	logFile = logrus.New()
	logGui  = logrus.New()
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
	//Setup GUI
	setupGui()
	defer g.Close()

	file, err := os.OpenFile("abv.log", os.O_CREATE|os.O_WRONLY, 0666)
	defer file.Close()
	if err == nil {
		logFile.Out = file
	} else {
		logFile.Info("Failed to log to file, using default stderr")
	}

	// Start Gui
	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		logFile.Fatal(err)
	}
}

func setupGui() {
	gui, err := gocui.NewGui(gocui.OutputNormal)
	g = gui
	if err != nil {
		logFile.Fatal(err)
	}

	g.SetManagerFunc(layout)
	g.Cursor = true

	if err := configureKeys(); err != nil {
		logFile.Fatalln(err)
	}
}

func parseInput(g *gocui.Gui, v *gocui.View) error {
	logFile.WithFields(logrus.Fields{
		"category": "userEntry",
		"action": "searchSelection",
	}).Info(v.Buffer())
	updatePopup(v.Buffer())
	togglePopup()
	clearInput()
	return nil
}

func updatePopup(name string) {
	g.Update(func(g *gocui.Gui) (err error) {
		v, err := g.View(popup)
		if err != nil {
			logFile.Error(err)
			return
		}

		drinks, err := SearchUntappdByName(name)
		if err != nil {
			logFile.Error(err)
			displayError(err)
			return
		}
		v.Clear()

		for _, drink := range drinks {
			fmt.Fprintf(v, "%s: %s\n", drink.Brand, drink.Name)
		}

		return
	})
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
