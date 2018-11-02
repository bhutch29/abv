package main

import (
	"fmt"
	"github.com/bhutch29/abv/model"
	"github.com/jroimartin/gocui"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
	"strings"
)

var (
	g       *gocui.Gui
	c       ModalController
	drinks  []model.Drink
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

	drinkFormat = "%s: %s\n"
)

var keys = []key{
	{"", gocui.KeyCtrlC, quit, "C-c", "quit"},
	{"", gocui.KeyCtrlE, testError, "C-e", "error"},
	{"", gocui.KeyCtrlI, setInputMode, "C-i", "stocking mode"},
	{"", gocui.KeyCtrlO, setOutputMode, "C-o", "serving mode"},
	{input, gocui.KeyEnter, parseInput, "Enter", "confirm"},
	{popup, gocui.KeyArrowUp, popupScrollUp, "Up", "scrollUp"},
	{popup, gocui.KeyCtrlK, popupScrollUp, "Up", "scrollUp"},
	{popup, gocui.KeyArrowDown, popupScrollDown, "Down", "scrollDown"},
	{popup, gocui.KeyCtrlJ, popupScrollDown, "Down", "scrollDown"},
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

	c, err = New()
	if err != nil {
		logFile.Error(err)
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
	text := strings.TrimSuffix(v.Buffer(), "\n")
	bc, err := strconv.Atoi(text)
	if err == nil {
		handleBarcode(bc)
	} else {
		handleSearch(logFile, text)
	}
	clearInput()
	return nil
}

func handleSearch(logFile *logrus.Logger, text string) {
	logFile.WithFields(logrus.Fields{
		"category": "userEntry",
		"entry":    text,
	}).Info("User searched for a drink")
	updatePopup(text)
	togglePopup()
}

func handleBarcode(bc int) {
	logGui.Info("Scanned barcode", bc)
	logFile.Info("Scanned barcode", bc)

	exists, err := c.HandleBarcode(bc)
	if err != nil {
		logGui.Error("Failed to search database for barcode", err)
		logFile.Error("Failed to search database for barcode", err)
	}
	if !exists {
		if c.GetMode() != stocking {
			logGui.Warn("Barcode not recognized while serving. Drink will not be recorded")
		} else {
			logGui.Info("Barcode not recognized. Please enter drink brand and name.")
			//TODO Change view state to indicate data entry needed? Maybe prevent data entry until this happens somehow?
			logFile.Info("Unknown barcode scanned", bc)
		}
	} else {
		logGui.Info("Barcode found") //TODO Return info on scanned beer!
		logFile.Info("Known barcode scanned", bc)
	}
}

func updatePopup(name string) {
	g.Update(func(g *gocui.Gui) (err error) {
		v, err := g.View(popup)
		if err != nil {
			logFile.Error(err)
			return
		}

		drinks, err = SearchUntappdByName(name)
		if err != nil {
			logFile.Error(err)
			displayError(err)
			return
		}
		v.Clear()

		for _, drink := range drinks {
			fmt.Fprintf(v, drinkFormat, drink.Brand, drink.Name)
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
	logGui.Info("You selected: " + line)

	d := findDrinkFromSelection(line)
	d.Barcode = c.LastBarcode()
	logGui.Info("Adding new drink", d)
	logFile.Info("Adding new drink", d)
	c.NewDrink(d)
	return err
}

func findDrinkFromSelection(line string) model.Drink {
	var d model.Drink
	var brand, name string
	fmt.Sscanf(line, drinkFormat, &brand, &name)
	for _, drink := range drinks {
		if drink.Brand == brand && drink.Name == name {
			d = drink
		}
	}
	return d
}

func setInputMode(g *gocui.Gui, v *gocui.View) error {
	c.SetMode(stocking)
	logGui.WithField("mode", stocking).Info("Changed Mode")
	logFile.WithField("mode", stocking).Info("Changed Mode")
	return nil
}

func setOutputMode(g *gocui.Gui, v *gocui.View) error {
	c.SetMode(serving)
	logGui.WithField("mode", serving).Info("Changed Mode")
	logFile.WithField("mode", serving).Info("Changed Mode")
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
