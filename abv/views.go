package main

import (
	"github.com/jroimartin/gocui"
	"fmt"
)

var stockDivisor = 2.5

const (
	inputHeight    = 4
	inputCursorPos = 12

	searchEntryHeight = 3
	searchCursorPos   = 4
)

func layout(g *gocui.Gui) (err error) {
	if err = makeLogPanel(); err != nil {
		logFile.Fatal(err)
		return
	}
	if err = makePromptPanel(); err != nil {
		logFile.Fatal(err)
		return
	}
	if err = makeInfoPanel(); err != nil {
		logFile.Fatal(err)
		return
	}
	if err = makeSelectOptionsPopup(); err != nil {
		logFile.Fatal(err)
		return
	}
	return
}

func makeLogPanel() error {
	maxX, maxY := g.Size()
	viewHeight := maxY - inputHeight
	logWidth := float64(maxX) - float64(maxX) / stockDivisor

	var x0, x1, y0, y1 int

	x0, x1 = 0, int(logWidth)
	y0, y1 = 0, viewHeight

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

func makeSelectOptionsPopup() error {
	maxX, maxY := g.Size()
	w := maxX / 2
	h := maxY / 4
	x0 := (maxX / 2) - (w / 2)
	y0 := (maxY / 2) - (h / 2)
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

func makeInfoPanel() error {
	maxX, maxY := g.Size()
	viewHeight := maxY - inputHeight
	infoStart := float64(maxX) - float64(maxX) / stockDivisor

	if v, err := g.SetView(info, int(infoStart), 0, maxX-2, viewHeight); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Stock"
		v.Wrap = true
		v.Clear()
		refreshInventory()
	}

	return nil
}


func makePromptPanel() error {
	maxX, maxY := g.Size()
	promptStartHeight := maxY - inputHeight
	promptDividerHeight := maxY - (inputHeight / 2)
	promptString := generateKeybindString()

	if v, err := g.SetView(prompt, 0, promptStartHeight, maxX, promptDividerHeight); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		fmt.Fprintf(v, promptString)
	}

	if v, err := g.SetView(input, inputCursorPos, promptDividerHeight, maxX, maxY); err != nil {
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

	if v, err := g.SetView(promptSymbol, 0, promptDividerHeight, inputCursorPos, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		updatePromptSymbol()
	}

	return nil
}
