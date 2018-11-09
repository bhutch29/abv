package main

import (
	"flag"
	"fmt"
	"github.com/bhutch29/abv/model"
	"github.com/jroimartin/gocui"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
	"log"
	"os/exec"
	"errors"
	aur "github.com/logrusorgru/aurora"
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
)

var keys = []key{
	{"", gocui.KeyCtrlI, setInputMode, "C-i", "stocking"},
	{"", gocui.KeyCtrlO, setOutputMode, "C-o", "serving"},
	{"", gocui.KeyCtrlC, quit, "C-c", "quit"},
	{input, gocui.KeyEnter, parseInput, "Enter", "confirm"},
	{search, gocui.KeyEnter, handleSearch, "Enter", "confirm"},
	{search, gocui.KeyCtrlZ, cancelSearch, "C-c", "cancel"},
	{popup, gocui.KeyArrowUp, popupScrollUp, "Up", "scrollUp"},
	{popup, gocui.KeyCtrlK, popupScrollUp, "Up", "scrollUp"},
	{popup, gocui.KeyArrowDown, popupScrollDown, "Down", "scrollDown"},
	{popup, gocui.KeyCtrlJ, popupScrollDown, "Down", "scrollDown"},
	{popup, gocui.KeyEnter, popupSelectItem, "Enter", "Select"},
	{errorView, gocui.KeyEsc, hideError, "Esc", "close error dialog"},
}

func main() {
	//Create Controller
	var err error
	if c, err = New(); err != nil {
		logFile.Error("Error creating controller: ", err)
	}

	//Setup loggers
	f := logrus.TextFormatter{}
	f.ForceColors = true
	f.DisableTimestamp = true
	f.DisableLevelTruncation = true
	logGui.Formatter = &f
	logGui.SetLevel(logrus.InfoLevel)

	file, err := os.OpenFile("abv.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		logFile.Out = file
	} else {
		logFile.Info("Failed to log to file, using default stderr")
	}
	defer file.Close()
	logFile.SetLevel(logrus.DebugLevel)

	//Command Line flags
	handleFlags()

	//Setup GUI
	setupGui()
	defer g.Close()

	// Start Gui
	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		logFile.Fatal(err)
	}
}

func handleFlags() {
	backup := flag.String("backup", "", "Backs up the sqlite database to specified file")
	reset := flag.Bool("reset", false, "Backs up the database to the working directory and wipes out the Input and Output tables")

	flag.Parse()

	if *backup != "" {
		backupDatabase(*backup)
		os.Exit(0)
	}

	if *reset {
		backupDatabase("backup.sqlite")
		if err := c.ClearInputOutputRecords(); err != nil {
			log.Print("Error clearing Input and Output records" + err.Error())
			logFile.Fatal(err)
		}
		os.Exit(0)
	}
}

func backupDatabase(destination string) {
	log.Print("Backup up database to " + destination)
	cmd := exec.Command("sqlite3", "abv.sqlite", ".backup " + destination)
	if err := cmd.Run(); err != nil {
		log.Print("Failed to backup database: " + err.Error())
		logFile.Fatal(err)
	}
}

func setupGui() {
	var err error
	g, err = gocui.NewGui(gocui.Output256)
	if err != nil {
		logFile.Fatal(err)
	}

	vd := viewDrawer{}
	g.SetManagerFunc(vd.layout)
	g.Cursor = true

	if err := configureKeys(); err != nil {
		logFile.Fatal(err)
	}
}

func refreshInventory() error {
	view, err := g.View(info)
	if err != nil {
		logGui.Error(err)
		logFile.Error(err)
	}
	view.Clear()
	inventory := c.GetInventory()
	for _, drink := range inventory {
		//TODO: Make this more robust to handle arbitrary length Brand and Name strings
		if len(drink.Name) < 30 {
			fmt.Fprintf(view, "%-35s%-30s%6d\n", drink.Brand, drink.Name, drink.Quantity)
		} else {
			fmt.Fprintf(view, "%-35s%-30s%6d\n", drink.Brand, drink.Name[:30], drink.Quantity)
			fmt.Fprintf(view, "%-35s%-30s%6s\n", "", drink.Name[30:], "")
		}
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
	logGui.Debug("Scanned barcode: ", bc)
	logFile.Debug("Scanned barcode: ", bc)

	exists, err := c.HandleBarcode(bc)
	if err != nil {
		logGui.Error("Failed to search database for barcode", err)
		logFile.Error("Failed to search database for barcode", err)
	}

	if !exists {
		handleNewBarcode(bc)
	}

	refreshInventory()
}

func handleNewBarcode(bc string) {
	if c.GetMode() != stocking {
		logGui.Warn("Barcode not recognized while serving. Drink will not be recorded")
		return
	}

	logGui.Info("Barcode not recognized. Please enter drink brand and name.")
	logFile.Info("Unknown barcode scanned: ", bc)
	clearView(popup)
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

func cancelSearch(g *gocui.Gui, v *gocui.View) error {
	togglePopup()
	logGui.Info("Canceled entering information for new barcode")
	logFile.Info("Canceled entering information for new barcode")
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
		fmt.Fprintf(v, "%s: %s\n", drink.Brand, drink.Name)
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
	}).Debug("User selected a beer")
	logGui.Debug("You selected: " + line)

	d, err := findDrinkFromSelection(line)
	if err != nil {
		logGui.Error(err)
		logFile.Error(err)
		return nil
	}

	d.Barcode = c.LastBarcode()

	logGui.Debug("Adding new drink", d)
	logFile.Debug("Adding new drink", d)

	if err = c.NewDrink(d); err != nil {
		logGui.Error(err)
		logFile.Error(err)
	}

	refreshInventory()

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
	return d, errors.New("Could not parse brand and drink name from selected text: " + line)
}

func setInputMode(g *gocui.Gui, v *gocui.View) error {
	c.SetMode(stocking)
	updatePromptSymbol()
	logGui.Infof("Changed to %s Mode", aur.Brown("Stocking"))
	logFile.WithField("mode", stocking).Info("Changed Mode")
	return nil
}

func setOutputMode(g *gocui.Gui, v *gocui.View) error {
	c.SetMode(serving)
	updatePromptSymbol()
	logGui.Infof("Changed to %s Mode", aur.Green("Serving"))
	logFile.WithField("mode", serving).Info("Changed Mode")
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
