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
	logView       = "Log"
	input         = "Input"
	info          = "Info"
	popup         = "Popup"
	prompt        = "Prompt"
	promptSymbol  = "PromptSymbol"
	errorView     = "Errors"
	search        = "Search"
	searchSymbol  = "SearchSymbol"
	searchOutline = "SearchOutline"

	drinkFormat = "%s: %s\n"
)

var keys = []key{
	{"", gocui.KeyCtrlC, quit, "C-c", "quit"},
	{"", gocui.KeyCtrlE, testError, "C-e", "error"},
	{"", gocui.KeyCtrlI, setInputMode, "C-i", "stocking mode"},
	{"", gocui.KeyCtrlO, setOutputMode, "C-o", "serving mode"},
	{input, gocui.KeyEnter, parseInput, "Enter", "confirm"},
	{search, gocui.KeyEnter, handleSearch, "Enter", "confirm"},
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
	logFile.SetLevel(logrus.DebugLevel)

	c, err = New()
	if err != nil {
		logFile.Error("Error creating controller: ", err)
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
		if handleBarcode(bc) {
			togglePopup()
		}
	} else {
		logGui.Warn("Barcode entry must be an integer")
		logFile.Warn("Non-integer barcode entered", text)
	}
	clearView(input)
	return nil
}

func handleSearch(g *gocui.Gui, v *gocui.View) error {
	s, _ := g.View(searchOutline)
	s.Title = ""
	clearView(search)
	text := v.Buffer()
	logFile.WithFields(logrus.Fields{
		"category": "userEntry",
		"entry":    text,
	}).Info("User searched for a drink")
	updatePopup(text)
	g.SetCurrentView(popup)
	p, _ := g.View(popup)
	p.Title = "Select desired drink..."
	return nil
}

func handleBarcode(bc int) bool {
	logGui.Info("Scanned barcode ", bc)
	logFile.Info("Scanned barcode ", bc)

	exists, err := c.HandleBarcode(bc)
	if err != nil {
		logGui.Error("Failed to search database for barcode", err)
		logFile.Error("Failed to search database for barcode", err)
	}
	if !exists {
		if c.GetMode() != stocking {
			logGui.Warn("Barcode not recognized while serving. Drink will not be recorded")
			return false
		}
		logGui.Info("Barcode not recognized. Please enter drink brand and name.")
		//TODO Change view state to indicate data entry needed? Maybe prevent data entry until this happens somehow?
		logFile.Info("Unknown barcode scanned", bc)
		return true
	}
	logGui.Info("Barcode found") //TODO Return info on scanned beer!
	logFile.Info("Known barcode scanned", bc)
	return false
}

func updatePopup(name string) {
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
	if err = c.NewDrink(d); err != nil {
		logGui.Error(err)
		logFile.Error(err)
	}
	return nil
}

func findDrinkFromSelection(line string) model.Drink {
	logFile.Debug("Finding drink from selected text: ", line)
	var d model.Drink
	s := strings.Split(line, ":")
	brand := s[0]
	name := strings.TrimSpace(s[1])
	logFile.Debug("Determined that brand = " + brand + " and name = " + name)
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
