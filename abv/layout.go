package main

import (
	"fmt"
	"github.com/jroimartin/gocui"
)

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
	return
}

// Define Prompt dimensions
const (
	inputHeight    = 4
	inputCursorPos = 4
	promptWidth    = 21
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

// Draw two panels on the bottom of the screen, one for input and one
// for keybinding information
func makePrompt(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	promptStartHeight := maxY - inputHeight
	promptDividerHeight := maxY - (inputHeight / 2)

	if v, err := g.SetView("Prompt", 0, promptStartHeight, promptWidth, promptDividerHeight); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		printPrompt(g)
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

func printPrompt(g *gocui.Gui) {
	promptString := "C-c: Quit"

	g.Update(func(g *gocui.Gui) error {
		input, err := g.View("Input")
		prompt, err := g.View("Prompt")
		if err != nil {
			return err
		}
		input.Clear()
		x, y := input.Cursor()
		input.MoveCursor(-x, -y, true)

		prompt.Clear()
		fmt.Fprintf(prompt, promptString)
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
