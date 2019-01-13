package main

import (
	"fmt"

	"github.com/jroimartin/gocui"
	aur "github.com/logrusorgru/aurora"
	"strconv"
	"strings"
)

type key struct {
	viewname  string
	key       interface{}
	handler   func(*gocui.Gui, *gocui.View) error
	shortkey  string
	shortname string
}

var keys []key

func initializekeys() {
	keys = []key{
		{"", gocui.KeyCtrlI, setInputMode, "Ctrl-i", "stocking"},
		{"", gocui.KeyCtrlO, setOutputMode, "Ctrl-o", "serving"},
		{"", gocui.KeyCtrlZ, undoLastKeyboardAction, "Ctrl-z", "undo"},
		{"", gocui.KeyCtrlR, redoLastKeyboardAction, "Ctrl-r", "redo"},
		{"", gocui.KeyCtrlC, quit, "Ctrl-c", "quit"},
		{"", gocui.KeyF1, setQuantity1, "F1", "single"},
		{"", gocui.KeyF4, setQuantity4, "F4", "four-pack"},
		{"", gocui.KeyF6, setQuantity6, "F6", "six-pack"},
		{"", gocui.KeyF12, setQuantity12, "F12", "twelve-pack"},
		{input, gocui.KeyArrowUp, scrollInventoryUp, "Up", "scroll up"},
		{input, gocui.KeyArrowDown, scrollInventoryDown, "Down", "scroll down"},
		{input, gocui.KeyEnter, parseInput, "Enter", "confirm"},
		{search, gocui.KeyEnter, handleSearch, "Enter", "confirm"},
		{search, gocui.KeyEsc, cancelSearch, "Ctrl-z", "cancel"},
		{popup, gocui.KeyEsc, cancelPopup, "Ctrl-z", "cancel"},
		{popup, gocui.KeyArrowUp, popupScrollUp, "Up", "scrollUp"},
		{popup, gocui.KeyCtrlK, popupScrollUp, "Up", "scrollUp"},
		{popup, gocui.KeyArrowDown, popupScrollDown, "Down", "scrollDown"},
		{popup, gocui.KeyCtrlJ, popupScrollDown, "Down", "scrollDown"},
		{popup, gocui.KeyEnter, popupSelectItem, "Enter", "Select"},
		{errorView, gocui.KeyEsc, hideError, "Esc", "close error dialog"},
	}
}

func generateKeybindString(quantity int) string {
	var result string
	for _, k := range keys {
		if k.viewname == "" || k.viewname == input{
			if getKeyQuantity(k.shortkey) == quantity {
				result = result + fmt.Sprintf("%s->%s ", aur.BgBlue(aur.Black(k.shortkey)), k.shortname)
			} else {
				result = result + fmt.Sprintf("%s->%s ", aur.BgGray(aur.Black(k.shortkey)), k.shortname)
			}
		}
	}
	return result
}

func getKeyQuantity(shortkey string) int {
	res, err := strconv.Atoi(strings.TrimPrefix(shortkey, "F"))
	if err != nil {
		return -1
	}
	return res
}

func configureKeys() error {
	for _, key := range keys {
		if err := g.SetKeybinding(key.viewname, key.key, gocui.ModNone, key.handler); err != nil {
			return err
		}
	}

	return nil
}
