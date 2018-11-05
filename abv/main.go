package main

import (
	"fmt"
	"github.com/bhutch29/abv/model"
	"github.com/jroimartin/gocui"
	"github.com/sirupsen/logrus"
	"os"
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
	{"", gocui.KeyCtrlQ, refreshInventory, "C-q", "get inventory"},
	{input, gocui.KeyEnter, parseInput, "Enter", "confirm"},
	{search, gocui.KeyEnter, handleSearch, "Enter", "confirm"},
	{popup, gocui.KeyArrowUp, popupScrollUp, "Up", "scrollUp"},
	{popup, gocui.KeyCtrlK, popupScrollUp, "Up", "scrollUp"},
	{popup, gocui.KeyArrowDown, popupScrollDown, "Down", "scrollDown"},
	{popup, gocui.KeyCtrlJ, popupScrollDown, "Down", "scrollDown"},
	{popup, gocui.KeyEnter, popupSelectItem, "Enter", "Select"},
	{errorView, gocui.KeyEsc, hideError, "Esc", "close error dialog"},
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
	if err == nil {
		logFile.Out = file
	} else {
		logFile.Info("Failed to log to file, using default stderr")
	}
	defer file.Close()
	logFile.SetLevel(logrus.DebugLevel)

	//Create Controller
	if c, err = New(); err != nil {
		logFile.Error("Error creating controller: ", err)
	}

	// Start Gui
	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		logFile.Fatal(err)
	}
}

func setupGui() {
	var err error
	g, err = gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		logFile.Fatal(err)
	}

	g.SetManagerFunc(layout)
	g.Cursor = true

	if err := configureKeys(); err != nil {
		logFile.Fatalln(err)
	}
}

func refreshInventory(g *gocui.Gui, v *gocui.View) error {
	view, _ := g.View(info)
	view.Clear()
	writeInventory()
	return nil
}

func writeInventory() error {
	view, _ := g.View(info)
	inventory := c.GetInventory()
	for _, drink := range inventory {
		//TODO: Make this more robust to handle arbitrary length Brand and Name strings
		fmt.Fprintf(view, "%-40s%-20s%6d\n", drink.Brand, drink.Name, drink.Quantity)
	}
	return nil
}

func parseInput(g *gocui.Gui, v *gocui.View) error {
	bc := strings.TrimSuffix(v.Buffer(), "\n")
	clearView(input)
	handleBarcodeEntry(bc)
	return nil
}

func handleBarcodeEntry(bc string) {
	logGui.Info("Scanned barcode: ", bc)
	logFile.Info("Scanned barcode: ", bc)

	exists, err := c.HandleBarcode(bc)
	if err != nil {
		logGui.Error("Failed to search database for barcode", err)
		logFile.Error("Failed to search database for barcode", err)
	}

	if !exists {
		handleNewBarcode(bc)
	}
}

func handleNewBarcode(bc string) {
	if c.GetMode() != stocking {
		logGui.Warn("Barcode not recognized while serving. Drink will not be recorded")
		return
	}

	logGui.Info("Barcode not recognized. Please enter drink brand and name.")
	logFile.Info("Unknown barcode scanned: ", bc)
	togglePopup()
}

func handleSearch(g *gocui.Gui, v *gocui.View) error {
	text := v.Buffer()

	logFile.WithFields(logrus.Fields{
		"category": "userEntry",
		"entry":    text,
	}).Info("User searched for a drink")

	setTitle(searchOutline, "")
	clearView(search)
	updatePopup(text)
	setTitle(popup, "Select desired drink...")
	return nil
}

func updatePopup(name string) {
	v, _ := g.View(popup)

	var err error
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

	g.SetCurrentView(popup)
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

	d, err := findDrinkFromSelection(line)
	if err != nil {
		logGui.Error(err)
		logFile.Error(err)
		return nil
	}

	d.Barcode = c.LastBarcode()

	logGui.Info("Adding new drink", d)
	logFile.Info("Adding new drink", d)

	if err = c.NewDrink(d); err != nil {
		logGui.Error(err)
		logFile.Error(err)
	}

	return nil
}

func findDrinkFromSelection(line string) (model.Drink, error) {
	logFile.Debug("Finding drink from selected text: ", line)
	var d model.Drink

	s := strings.Split(line, ":")
	brand := s[0]
	name := strings.TrimSpace(s[1])

	logFile.Debug("Determined that brand = " + brand + " and name = " + name)

	for _, drink := range drinks {
		if drink.Brand == brand && drink.Name == name {
			return drink, nil
		}
	}
	return d, fmt.Errorf("Could not parse brand and drink name from selected text: " + line)
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
