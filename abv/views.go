package main

import (
	"fmt"

	"github.com/jroimartin/gocui"
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

const stockDivisor = 2

const (
	inputHeight    = 4
	inputCursorPos = 12

	searchEntryHeight = 3
	searchCursorPos   = 4
)

type viewDrawer struct {
	maxX int
	maxY int
}

func (vd *viewDrawer) layout(g *gocui.Gui) (err error) {
	vd.maxX, vd.maxY = g.Size()
	if err = vd.makeLogPanel(); err != nil {
		logFile.Fatal(err)
		return
	}
	if err = vd.makePromptPanels(); err != nil {
		logFile.Fatal(err)
		return
	}
	if err = vd.makeInfoPanel(); err != nil {
		logFile.Fatal(err)
		return
	}
	if err = vd.makeSelectOptionsPopup(); err != nil {
		logFile.Fatal(err)
		return
	}
	return
}

func (vd *viewDrawer) makeLogPanel() error {
	viewHeight := vd.maxY - inputHeight
	logWidth := float64(vd.maxX) - float64(vd.maxX)/stockDivisor

	x0 := 0
	x1 := int(logWidth)
	y0 := 0
	y1 := viewHeight

	if v, err := g.SetView(logView, x0, y0, x1, y1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Wrap = true
		v.Title = "Log"
		v.Autoscroll = true
		logGui.Out = v
	}

	return nil
}

func (vd *viewDrawer) makeSelectOptionsPopup() error {
	w := vd.maxX / 2
	h := vd.maxY / 4

	x0 := (vd.maxX / 2) - (w / 2)
	y0 := (vd.maxY / 2) - (h / 2)
	x1 := x0 + w
	y1 := y0 + h

	if v, err := g.SetView(popup, x0, y0, x1, y1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Frame = true
		v.Highlight = true
		//TODO: Set Selected FG and BG colors
		g.SetViewOnBottom(popup)
	}

	if v, err := g.SetView(search, x0+searchCursorPos, y0+h+1, x1, y0+h+searchEntryHeight); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Editable = true
		v.Wrap = false
		v.Frame = false
		v.Editor = gocui.EditorFunc(promptEditor)
		g.SetViewOnBottom(search)
	}

	if v, err := g.SetView(searchSymbol, x0, y0+h+1, x0+searchCursorPos, y0+h+searchEntryHeight); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		fmt.Fprintf(v, ">>")
		g.SetViewOnBottom(searchSymbol)
	}

	if v, err := g.SetView(searchOutline, x0, y0+h, x1, y0+h+searchEntryHeight); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = true
		g.SetViewOnBottom(searchOutline)
	}

	return nil
}

func (vd *viewDrawer) makeInfoPanel() error {
	viewHeight := vd.maxY - inputHeight
	infoStart := float64(vd.maxX) - float64(vd.maxX)/stockDivisor

	if v, err := g.SetView(info, int(infoStart), 0, vd.maxX-2, viewHeight); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Stock"
		v.Wrap = true
		refreshInventory()
	}

	return nil
}

func (vd *viewDrawer) makePromptPanels() error {
	promptStartHeight := vd.maxY - inputHeight
	promptDividerHeight := vd.maxY - (inputHeight / 2)
	promptString := generateKeybindString()

	if v, err := g.SetView(prompt, 0, promptStartHeight, vd.maxX, promptDividerHeight); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		fmt.Fprintf(v, promptString)
	}

	if v, err := g.SetView(input, inputCursorPos, promptDividerHeight, vd.maxX, vd.maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		v.Editable = true
		v.Wrap = false
		v.Editor = gocui.EditorFunc(promptEditor)
		if _, err := g.SetCurrentView(input); err != nil {
			return err
		}
	}

	if v, err := g.SetView(promptSymbol, 0, promptDividerHeight, inputCursorPos, vd.maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		updatePromptSymbol()
	}

	return nil
}
