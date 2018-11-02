package main

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/sirupsen/logrus"
	"os"
)

var (
	g       *gocui.Gui
	c       ModalController
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
	{"", gocui.KeyCtrlE, testError, "C-e", "error"},
	{input, gocui.KeyEnter, parseInput, "Enter", "confirm"},
	{popup, gocui.KeyArrowUp, popupScrollUp, "Up", "scrollUp"},
	{popup, gocui.KeyArrowDown, popupScrollDown, "Down", "scrollDown"},
	{popup, gocui.KeyEnter, popupSelectItem, "Enter", "Select"},
}

func testError(g *gocui.Gui, v *gocui.View) error {
	logGui.WithFields(logrus.Fields{
		"Category":    "Test",
		"CurrentView": v.Name(),
	}).Error("This is an example error for testing purposes")
	logFile.WithFields(logrus.Fields{
		"Category":    "Test",
		"CurrentView": v.Name(),
	}).Error("This is an example error for testing purposes")
	return nil
}

func main() {
	//Setup GUI
	setupGui()
	defer g.Close()

	//Setup loggers
	f := logrus.TextFormatter{}
	f.ForceColors = true
	f.DisableTimestamp = true
	f.DisableLevelTruncation = true
	logGui.Formatter = &f

	file, err := os.OpenFile("abv.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
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
		"entry":    v.Buffer(),
	}).Info("User searched for a drink")
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

func popupSelectItem(g *gocui.Gui, v *gocui.View) error {
	line, err := getViewLine(v)
	togglePopup()
	resetViewCursor(v)
	logFile.WithFields(logrus.Fields{
		"category": "userEntry",
		"entry":    line,
	}).Info("User selected a beer")

	//TODO Do something with selected value
	// c, err := New()
	// if err != nil {
	// 	logFile.Error(err)
	// 	logGui.Error(err)
	// }
	logGui.Info("You selected: " + line)
	return err
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
