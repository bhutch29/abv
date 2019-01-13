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
	file := stderrDest(logFile)
	if file != nil {
		defer file.Close()
	}

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

func refreshInventory() error {
	view, err := g.View(info)
	if err != nil {
		logAllError(err)
	}
	view.Clear()
	inventory := c.GetInventorySorted([]string{"brand", "name"})
	total := c.GetInventoryTotalQuantity()
	variety := c.GetInventoryTotalVariety()
	fmt.Fprintf(view, "Total Beers: %d      Total Varieties: %d\n\n", total, variety)
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

func parseInput(g *gocui.Gui, v *gocui.View) error {
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

func cancelPopup(g *gocui.Gui, v *gocui.View) error {
	togglePopup()
	logAllInfo("Canceled selecting beer from list")
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

func shortenType(in string) string {
	split := strings.SplitN(in, " - ", 2)
	return split[0]
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
		setQuantity1(g, v)
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

func scrollInventoryUp(g *gocui.Gui, v *gocui.View) error {
	vi, _ := g.View(info)
	scrollView(vi, -1)
	return nil
}

func scrollInventoryDown(g *gocui.Gui, v *gocui.View) error {
	vi, _ := g.View(info)
	scrollView(vi, 1)
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

		v, _ := g.View(prompt)
		v.Clear()
		fmt.Fprintf(v, generateKeybindString(q))
	}
}

func setQuantity1(g *gocui.Gui, v *gocui.View) error {
	trySetQuantity(1)
	return nil
}

func setQuantity4(g *gocui.Gui, v *gocui.View) error {
	trySetQuantity(4)
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
