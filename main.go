package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/bhutch29/abv/cache"
	"github.com/bhutch29/abv/config"
	"github.com/bhutch29/abv/model"
	"github.com/jroimartin/gocui"
	aur "github.com/logrusorgru/aurora"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/text/unicode/norm"
)

var (
	g        *gocui.Gui
	c        ModalController
	drinks   []model.Drink
	quantity int
	conf     *viper.Viper
	version  = "undefined"
)

func init() {
	quantity = 1

	initializekeys()

	//Setup loggers
	f := logrus.TextFormatter{}
	f.ForceColors = true
	f.DisableTimestamp = true
	f.DisableLevelTruncation = true
	logGui.Formatter = &f
	logGui.SetLevel(logrus.InfoLevel)
	logFile.SetLevel(logrus.DebugLevel)
}

func main() {
	// Save the state of terminal, so we can restore it after a panic
	fd := int(os.Stdout.Fd())
	oldState, err := terminal.GetState(fd)
	if err == nil {
		defer terminal.Restore(fd, oldState)
	}

	//Get Configuration
	if conf, err = config.New(); err != nil {
		log.Fatal("Error getting configuration info: ", err)
	}

	// Redirect stderr to log file
	file := redirectStderr(logFile)
	defer file.Close()

	//Create Controller
	if c, err = New(); err != nil {
		logFile.Fatal("Error creating controller: ", err)
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
		//TODO: backup to configPath
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

// refreshInventory displays the sorted inventory in the inventory view.
//
// The inventory is sorted first by drink brand, then by drink name.
func refreshInventory() error {
	view, err := g.View(info)
	if err != nil {
		logAllError(err)
	}
	view.Clear()
	inventory := c.GetInventorySorted([]string{"brand", "name"})
	total := c.GetInventoryTotalQuantity()
	variety := c.GetInventoryTotalVariety()
	fmt.Fprintf(view, "Total Drinks: %d     Total Varieties: %d\n\n", total, variety)
	for _, drink := range inventory {
		//TODO: Make this more robust to handle arbitrary length Brand and Name strings
		nfcBytes := norm.NFC.Bytes([]byte(drink.Name))
		nfcRunes := []rune(string(nfcBytes))
		visualLen := len(nfcRunes)
		if visualLen < 30 {
			fmt.Fprintf(view, "%-4d%-35s%-30s\n", drink.Quantity, drink.Brand, drink.Name)
		} else {
			const wsPad = "                                       " // strings.Repeat(" ", 39)
			fmt.Fprintf(view, "%-4d%-35s%-30s...\n", drink.Quantity, drink.Brand, string(nfcRunes[:30]))
			fmt.Fprintf(view, "%s...%s\n", wsPad, string(nfcRunes[30:]))
		}
	}
	return nil
}

// parseInput handles all input to the user interface and determines whether
// it should be handled as a barcode or as a predefined action. (i.e. undo/redo)
func parseInput(_ *gocui.Gui, v *gocui.View) error {
	bc := strings.TrimSuffix(v.Buffer(), "\n")
	clearView(input)
	if bc == "" {
		return nil
	}

	id, barcode := parseIDFromBarcode(bc)
	undoCode := conf.GetString("undoBarcode")
	redoCode := conf.GetString("redoBarcode")
	if barcode == undoCode {
		c.Undo(id)
		refreshInventory()
	} else if barcode == redoCode {
		c.Redo(id)
		refreshInventory()
	} else {
		handleBarcodeEntry(id, barcode)
	}
	return nil
}

// parseIDFromBarcode returns the input device ID from a line of text.
//
// If the input device is a keyboard, there is no corresponding ID. However
// if the input device is a scanner, the scanner is assumed to add a prefix
// "{c}_" where {c} is a single byte character unique to the scanner.
func parseIDFromBarcode(bc string) (string, string) {
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

// handleNewBarcode determines whether an unrecognized barcode should initiate
// the creation of a new drink model.
//
// In serving mode, the attempt is logged and no action is taken.
func handleNewBarcode() {
	if c.GetMode() != stocking {
		logGui.Warn("Barcode not recognized while serving. Drink will not be recorded")
		return
	}

	logAllInfo("Barcode not recognized. Please enter drink brand and name.")
	clearView(popup)
	togglePopup()
}

func handleSearch(_ *gocui.Gui, v *gocui.View) error {
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

func cancelSearch(_ *gocui.Gui, _ *gocui.View) error {
	togglePopup()
	logAllInfo("Canceled entering information for new barcode")
	return nil
}

// cancelPopup hides the drink selection popup and returns to the normal
// user interface.
func cancelPopup(_ *gocui.Gui, _ *gocui.View) error {
	togglePopup()
	logAllInfo("Canceled selecting drink from list")
	return nil
}

// updatePopup produces a popup to select the desired drink. It is populated
// with all of the results that match the provided query to the Untappd service.
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
		fmt.Fprintf(v, "%s:: %s\n", drink.Brand, drink.Name)
	}

	g.SetCurrentView(popup)
	return
}

// popupSelectItem takes the user's selected drink, creates a new drink model
// from the Untappd response, caches a brand image if not already cached,
// and finally refreshes the displayed inventory.
func popupSelectItem(_ *gocui.Gui, v *gocui.View) error {
	line, err := getViewLine(v)
	if err != nil {
		logAllError(err)
		return nil
	}

	togglePopup()
	resetViewCursor(v)

	logFile.WithFields(logrus.Fields{
		"category": "userEntry",
		"entry":    line,
	}).Debug("User selected a drink")
	logGui.Debug("You selected: " + line)

	d, err := findDrinkFromSelection(line)
	if err != nil {
		logAllError(err)
		return nil
	}

	err = cache.Image(d.Logo)
	if err != nil {
		logAllError("Failed HTTP request while caching image for drink: ", d.Brand, " ", d.Name)
	}

	d.Barcode = c.LastBarcode()
	d.Shorttype = shortenType(d.Type)
	id := c.LastID()

	logAllDebug("Adding new drink", d)

	if err = c.NewDrink(id, d, quantity); err != nil {
		logAllError(err)
	}

	refreshInventory()

	return nil
}

// shortenType produces an abbreviated drink type by truncating anything
// following the first hyphen.
//
// For example, "IPA - Double" would become just "IPA"
func shortenType(in string) string {
	split := strings.SplitN(in, " - ", 2)
	return split[0]
}

// findDrinkFromSelection takes the user's drink selection and associates it
// with the corresponding drink as queried from Untappd.
func findDrinkFromSelection(line string) (model.Drink, error) {
	logFile.Debug("Finding drink from selected text: ", line)
	var d model.Drink

	s := strings.Split(line, "::")
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

// setInputMode prepares the modal controller for stocking mode.
func setInputMode(_ *gocui.Gui, _ *gocui.View) error {
	if m := c.GetMode(); m != stocking {
		c.SetMode(stocking)
		updatePromptSymbol()
		logGui.Infof("Changed to %s Mode", aur.Brown("Stocking"))
		logFile.WithField("mode", stocking).Info("Changed Mode")
	}
	return nil
}

// setOutputMode prepares the modal controller for serving mode.
func setOutputMode(g *gocui.Gui, v *gocui.View) error {
	if m := c.GetMode(); m != serving {
		c.SetMode(serving)
		setQuantity1(g, v)
		updatePromptSymbol()
		logGui.Infof("Changed to %s Mode", aur.Green("Serving"))
		logFile.WithField("mode", serving).Info("Changed Mode")
	}
	return nil
}

// undoLastKeyboardAction reverts the previously performed action by the keyboard.
func undoLastKeyboardAction(_ *gocui.Gui, _ *gocui.View) error {
	c.Undo("")
	refreshInventory()
	return nil
}

// redoLastKeyboardAction performs the previously reverted action by the keyboard.
func redoLastKeyboardAction(_ *gocui.Gui, _ *gocui.View) error {
	c.Redo("")
	refreshInventory()
	return nil
}

// scrollInventoryUp retreats the cursor to the previous row in the inventory view.
//
// If no previous row exists, no action is taken.
func scrollInventoryUp(g *gocui.Gui, _ *gocui.View) error {
	vi, _ := g.View(info)
	scrollView(vi, -1)
	return nil
}

// scrollInventoryDown advances the cursor to the next row in the inventory view.
//
// If no next row exists, no action is taken.
func scrollInventoryDown(g *gocui.Gui, _ *gocui.View) error {
	vi, _ := g.View(info)
	scrollView(vi, 1)
	return nil
}

// trySetQuantity sets the quantity-per-scan to the given quantity q.
//
// If the user is in serving mode, the quantity is set to 1.
func trySetQuantity(q int) {
	if q != 1 && c.GetMode() != stocking {
		logAllInfo("Serving of multiple drinks at once is not supported")
		return
	}
	if q != quantity {
		quantity = q
		logAllInfo("Quantity of drinks per scan changed to ", quantity)

		v, _ := g.View(prompt)
		v.Clear()
		fmt.Fprintf(v, generateKeybindString(q))
	}
}

// setQuantity1 prepares the controller for either the scanning or serving
// of single beverages.
func setQuantity1(_ *gocui.Gui, _ *gocui.View) error {
	trySetQuantity(1)
	return nil
}

// setQuantity4 prepares the controller for scanning of 4-packs.
func setQuantity4(_ *gocui.Gui, _ *gocui.View) error {
	trySetQuantity(4)
	return nil
}

// setQuantity6 prepares the controller for scanning of 6-packs.
func setQuantity6(_ *gocui.Gui, _ *gocui.View) error {
	trySetQuantity(6)
	return nil
}

// setQuantity12 prepares the controller for scanning of 12-packs.
func setQuantity12(_ *gocui.Gui, _ *gocui.View) error {
	trySetQuantity(12)
	return nil
}

// quit provides a clean escape from the main gocui loop.
func quit(_ *gocui.Gui, _ *gocui.View) error {
	return gocui.ErrQuit
}
