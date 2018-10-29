package main

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"log"
)

var (
	g *gocui.Gui

	// View names
	mainView     = "Main"
	input        = "Input"
	info         = "Info"
	popup        = "Popup"
	prompt       = "Prompt"
	promptSymbol = "PromptSymbol"
	errorView    = "Errors"
)

// List of project level todos:
// TODO: Refactor: should any of these functions be methods? are they in the right files?
// TODO: Refactor colors
// TODO: Handle errors more consistently instead of just passing them up. What level should print errors for user?
// TODO: Set operation into 3 "modes": Stocking, Serving, and Auditing(Admin mode). How to manage these modes?
// TODO: Change "Stocking" mode entry point to be via barcode scanner (instead of entering name) and only ask for name if barcode is not found
// TODO: Organize the GUI into useful elements. Thoughts: Log (for displaying a feed of user info), Keybindings? (For displaying keybindings for current view), Input (For inputting text), Mode (For making it very apparent what mode we are in)
// TODO: Write log to text file for debugging?
// TODO: Stocking mode: Display number of scanned drink in inventory when scanned
// TODO: Serving mode: Display appropriate information when scanning drink out
// TODO: Admin mode: deleteDrink action, backup and clear inventory action

var keys []key = []key{
	key{"", gocui.KeyCtrlC, quit, "C-c", "quit"},
	key{input, gocui.KeyEnter, parseInput, "Enter", "confirm"},
	key{popup, gocui.KeyArrowUp, popupScrollUp, "Up", "scrollUp"},
	key{popup, gocui.KeyArrowDown, popupScrollDown, "Down", "scrollDown"},
	key{popup, gocui.KeyEnter, popupSelectItem, "Enter", "Select"},
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
	v, err := g.View(mainView)
	if err != nil {
		return err
	}
	fmt.Fprintln(v, s)
	return nil
}

func parseInput(g *gocui.Gui, v *gocui.View) error {
	vn, err := g.View(mainView)
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
