package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/bhutch29/abv/model"
	"github.com/jroimartin/gocui"
	aur "github.com/logrusorgru/aurora"
	"github.com/sirupsen/logrus"
)

var (
	g      *gocui.Gui
	c      ModalController
	drinks []model.Drink
	quantity int
	undoString = "65748392"
	redoString = "9384756"
	version = "undefined"
)

func main() {
	// Redirect stderr to log file
	file := redirectStderr(logFile)
	defer file.Close()

	//Create Controller
	var err error
	if c, err = New(); err != nil {
		logFile.Error("Error creating controller: ", err)
	}

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
	ver := flag.Bool("version", false, "Prints the version")
	verbose := flag.Bool("v", false, "Increases the logging verbosity in the GUI")

	flag.Parse()

	if *ver {
		fmt.Println(version)
		os.Exit(0)
	}

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

	if *verbose {
		logGui.SetLevel(logrus.DebugLevel)
	}
}

func backupDatabase(destination string) {
	log.Print("Backup up database to " + destination)
	cmd := exec.Command("sqlite3", "abv.sqlite", ".backup "+destination)
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
		logAllError(err)
	}
	view.Clear()
	inventory := c.GetInventory()
	for _, drink := range inventory {
		//TODO: Make this more robust to handle arbitrary length Brand and Name strings
		if len(drink.Name) < 30 {
			fmt.Fprintf(view, "%-4d%-35s%-30s\n", drink.Quantity, drink.Brand, drink.Name)
		} else {
			fmt.Fprintf(view, "%-4d%-35s%-30s\n", drink.Quantity, drink.Brand, drink.Name[:30])
			fmt.Fprintf(view, "%-4s%-35s%-30s\n", "", drink.Name[30:], "")
		}
	}
	return nil
}

func parseInput(g *gocui.Gui, v *gocui.View) error {
	bc := strings.TrimSuffix(v.Buffer(), "\n")
	clearView(input)
	if bc == "" {
		return nil
	}

	id, barcode := parseIDFromBarcode(bc)
	if barcode == undoString {
		c.Undo(id)
		refreshInventory()
	} else if barcode == redoString {
		c.Redo(id)
		refreshInventory()
	} else {
		handleBarcodeEntry(id, barcode)
	}
	return nil
}

func  parseIDFromBarcode(bc string) (string, string) {
	// If the second character is an _, treat the first character as a scanner ID and the rest of the input as a barcode
	if len(bc) == 1 {
		return "", bc
	}
	if bc[1] == []byte("_")[0] {
		return string(bc[0]), bc[2:]
	}
	return "", bc
}

func handleBarcodeEntry(id string, bc string) {
	logAllDebug("Scanned barcode: ", bc, " with ID=", id)
	exists, err := c.HandleBarcode(id, bc, quantity)
	if err != nil {
		logAllError("Failed to search database for barcode", err)
		return
	}

	if !exists {
		handleNewBarcode()
	}

	refreshInventory()
}

func handleNewBarcode() {
	if c.GetMode() != stocking {
		logGui.Warn("Barcode not recognized while serving. Drink will not be recorded")
		return
	}

	logAllInfo("Barcode not recognized. Please enter drink brand and name.")
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
	logAllInfo("Canceled entering information for new barcode")
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
		logAllError(err)
		return nil
	}

	d.Barcode = c.LastBarcode()
	id := c.LastID()

	logAllDebug("Adding new drink", d)

	if err = c.NewDrink(id, d, quantity); err != nil {
		logAllError(err)
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
	if m := c.GetMode(); m != stocking {
		c.SetMode(stocking)
		updatePromptSymbol()
		logGui.Infof("Changed to %s Mode", aur.Brown("Stocking"))
		logFile.WithField("mode", stocking).Info("Changed Mode")
	}
	return nil
}

func setOutputMode(g *gocui.Gui, v *gocui.View) error {
	if m := c.GetMode(); m != serving {
		c.SetMode(serving)
		updatePromptSymbol()
		logGui.Infof("Changed to %s Mode", aur.Green("Serving"))
		logFile.WithField("mode", serving).Info("Changed Mode")
	}
	return nil
}

func undoLastKeyboardAction(g *gocui.Gui, v *gocui.View) error {
	c.Undo("")
	refreshInventory()
	return nil
}

func redoLastKeyboardAction(g *gocui.Gui, v *gocui.View) error {
	c.Redo("")
	refreshInventory()
	return nil
}

func trySetQuantity(q int) {
	if q != 1 && c.GetMode() != stocking {
		logAllInfo("Serving of multiple beers at once is not supported")
		return
	}
	if q != quantity {
		quantity = q
		logAllInfo("Quantity of drinks per scan changed to ", quantity)
	}
}

func setQuantity1(g *gocui.Gui, v *gocui.View) error {
	trySetQuantity(1)
	return nil
}

func setQuantity6(g *gocui.Gui, v *gocui.View) error {
	trySetQuantity(6)
	return nil
}

func setQuantity12(g *gocui.Gui, v *gocui.View) error {
	trySetQuantity(12)
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
