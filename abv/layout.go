package main

import (
	"fmt"

	"github.com/jroimartin/gocui"
)

var popupDisplayed = false

func layout(g *gocui.Gui) (err error) {
	if err = makePrompt(g); err != nil {
		return
	}
	if err = makeMainPanels(g); err != nil {
		return
	}
	if err = makeInfoPanel(g); err != nil {
		return
	}
	if err = makeSelectOptionsPopup(g); err != nil {
		return
	}
	return
}

// Define Prompt dimensions
const (
	inputHeight    = 4
	inputCursorPos = 4
)

func makeMainPanels(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	viewHeight := maxY - inputHeight
	branchViewWidth := (maxX / 5) * 2

	var x0, x1, y0, y1 int

	x0, x1 = 0, branchViewWidth*2
	y0, y1 = 0, viewHeight

	if v, err := g.SetView("Main", x0, y0, x1, y1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Wrap = true
	}

	return nil
}

func makeSelectOptionsPopup(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	w := maxX / 2
	h := maxY / 4
	x0 := (maxX / 2) - (w / 2)
	y0 := (maxY / 2) - (h / 2)
	x1 := x0 + w
	y1 := y0 + h

	if v, err := g.SetView("Popup", x0, y0, x1, y1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Title = "Choose desired drink..."
		v.Frame = true
		v.Highlight = true
		g.SetViewOnBottom("Popup")
	}
	return nil
}

func togglePopup(g *gocui.Gui, v *gocui.View) error {
	vn := "Popup"

	if !popupDisplayed {
		g.SetViewOnTop(vn)
		g.SetCurrentView(vn)
	} else {
		g.SetViewOnBottom(vn)
		g.SetCurrentView("Input")
	}

	popupDisplayed = !popupDisplayed
	return nil
}

// Draw two panels on the bottom of the screen, one for input and one
// for keybinding information
func makePrompt(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	promptStartHeight := maxY - inputHeight
	promptDividerHeight := maxY - (inputHeight / 2)
	promptString := generateKeybindString()

	if v, err := g.SetView("Prompt", 0, promptStartHeight, maxY, promptDividerHeight); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		fmt.Fprintf(v, promptString)
	}

	if v, err := g.SetView("Input", inputCursorPos, promptDividerHeight, maxX, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		v.Editable = true
		v.Wrap = false
		v.Editor = gocui.EditorFunc(promptEditor)
		if _, err := g.SetCurrentView("Input"); err != nil {
			return err
		}
	}

	if v, err := g.SetView("PromptSymbol", 0, promptDividerHeight, inputCursorPos, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		fmt.Fprintf(v, ">>")
	}
	return nil
}

func makeInfoPanel(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	viewHeight := maxY - inputHeight
	branchViewWidth := (maxX / 5) * 2

	if v, err := g.SetView("Info", branchViewWidth*2, 0, maxX-2, viewHeight); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Info"
	}
	return nil
}

func clearInput(g *gocui.Gui) {
	g.Update(func(g *gocui.Gui) error {
		input, err := g.View("Input")
		if err != nil {
			return err
		}
		input.Clear()
		x, y := input.Cursor()
		input.MoveCursor(-x, -y, true)

		return nil
	})
}

func promptEditor(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	if ch != 0 && mod == 0 {
		v.EditWrite(ch)
		return
	}

	switch key {
	case gocui.KeySpace:
		v.EditWrite(' ')
	case gocui.KeyBackspace, gocui.KeyBackspace2:
		v.EditDelete(true)
	case gocui.KeyDelete:
		v.EditDelete(false)
	case gocui.KeyInsert:
		v.Overwrite = !v.Overwrite
	case gocui.KeyArrowDown:
		_ = v.SetCursor(len(v.Buffer())-1, 0)
	case gocui.KeyArrowUp:
		v.MoveCursor(0, -1, false)
	case gocui.KeyArrowLeft:
		v.MoveCursor(-1, 0, false)
	case gocui.KeyArrowRight:
		v.MoveCursor(1, 0, false)
	}
}

func displayError(g *gocui.Gui, e error) error {
	maxX, maxY := g.Size()
	x0 := maxX / 6
	x1 := maxY / 6
	y0 := 5 * (maxX / 6)
	y1 := 5 * (maxY / 6)

	if v, err := g.SetView("errors", x0, y0, x1, y1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		// Settings
		v.Title = "ERROR"
		v.Frame = true
		v.Wrap = true
		v.Autoscroll = true
		v.BgColor = gocui.ColorRed
		v.FgColor = gocui.ColorWhite

		// Content
		v.Clear()
		fmt.Fprintln(v, e.Error())

		// Send to forground
		g.SetCurrentView(v.Name())
	}

	return nil
}

func hideError(g *gocui.Gui) {
	g.DeleteView("errors")
}
