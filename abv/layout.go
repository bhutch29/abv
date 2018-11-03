package main

import (
	"fmt"

	"github.com/jroimartin/gocui"
)

var popupDisplayed = false

const (
	inputHeight    = 4
	inputCursorPos = 4

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
	branchViewWidth := (maxX / 5) * 2

	var x0, x1, y0, y1 int

	x0, x1 = 0, branchViewWidth*2
	y0, y1 = 0, viewHeight

	if v, err := g.SetView(logView, x0, y0, x1, y1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Wrap = true
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

func togglePopup() {
	if !popupDisplayed {
		g.SetViewOnTop(popup)
		g.SetViewOnTop(searchOutline)
		g.SetViewOnTop(searchSymbol)
		g.SetViewOnTop(search)
		g.SetCurrentView(search)
		setTitle(searchOutline, "Enter brewery and beer name...")
	} else {
		setTitle(popup, "")
		g.SetViewOnBottom(popup)
		g.SetViewOnBottom(searchSymbol)
		g.SetViewOnBottom(searchOutline)
		g.SetViewOnBottom(search)
		g.SetCurrentView(input)
	}

	popupDisplayed = !popupDisplayed
}

// Draw two panels on the bottom of the screen, one for input and one
// for keybinding information
func makePromptPanel() error {
	maxX, maxY := g.Size()
	promptStartHeight := maxY - inputHeight
	promptDividerHeight := maxY - (inputHeight / 2)
	promptString := generateKeybindString()

	if v, err := g.SetView(prompt, 0, promptStartHeight, maxY, promptDividerHeight); err != nil {
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
		fmt.Fprintf(v, ">>")
	}

	return nil
}

func makeInfoPanel() error {
	maxX, maxY := g.Size()
	viewHeight := maxY - inputHeight
	branchViewWidth := (maxX / 5) * 2

	if v, err := g.SetView(info, branchViewWidth*2, 0, maxX-2, viewHeight); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Info"
	}
	return nil
}

func clearView(view string) {
	g.Update(func(g *gocui.Gui) error {
		v, err := g.View(view)
		if err != nil {
			return err
		}
		v.Clear()
		x, y := v.Cursor()
		v.MoveCursor(-x, -y, true)

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

	}
}

func displayError(e error) error {
	maxX, maxY := g.Size()
	x0 := maxX / 6
	y0 := maxY / 6
	x1 := 5 * (maxX / 6)
	y1 := 5 * (maxY / 6)

	if v, err := g.SetView(errorView, x0, y0, x1, y1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Title = "ERROR"
		v.Frame = true
		v.Wrap = true
		v.Autoscroll = true
		v.BgColor = gocui.ColorRed
		v.FgColor = gocui.ColorWhite

		v.Clear()
		fmt.Fprintln(v, e.Error())
		g.SetCurrentView(v.Name())
	}

	return nil
}

func hideError(g *gocui.Gui, v *gocui.View) error {
	g.DeleteView(errorView)
	return nil
}

func popupScrollUp(g *gocui.Gui, v *gocui.View) error {
	err := moveViewCursorUp(v)
	if err != nil {
		logFile.Error(err)
		logGui.Error(err)
	}
	return nil
}

func popupScrollDown(g *gocui.Gui, v *gocui.View) error {
	err := moveViewCursorDown(v, false)
	if err != nil {
		logFile.Error(err)
		logGui.Error(err)
	}
	return err
}

func moveViewCursorDown(v *gocui.View, allowEmpty bool) error {
	cx, cy := v.Cursor()
	ox, oy := v.Origin()
	nextLine, err := getNextViewLine(v)
	if err != nil {
		return err
	}
	if !allowEmpty && nextLine == "" {
		return nil
	}
	if err := v.SetCursor(cx, cy+1); err != nil {
		if err := v.SetOrigin(ox, oy+1); err != nil {
			return err
		}
	}
	return nil
}

func moveViewCursorUp(v *gocui.View) error {
	ox, oy := v.Origin()
	cx, cy := v.Cursor()
	if cy >= 0 && oy > 0 {
		if err := v.SetCursor(cx, cy-1); err != nil {
			if err := v.SetOrigin(ox, oy-1); err != nil {
				return err
			}
		}
	}
	return nil
}

func getViewLine(v *gocui.View) (string, error) {
	var l string
	var err error

	_, cy := v.Cursor()
	if l, err = v.Line(cy); err != nil {
		l = ""
	}

	return l, err
}

func getNextViewLine(v *gocui.View) (string, error) {
	var l string
	var err error

	_, cy := v.Cursor()
	if l, err = v.Line(cy + 1); err != nil {
		l = ""
	}

	return l, err
}

func resetViewCursor(v *gocui.View) error {
	ox, _ := v.Origin()
	cx, _ := v.Cursor()
	if err := v.SetCursor(ox, 0); err != nil {
		if err := v.SetOrigin(cx, 0); err != nil {
			return err
		}
	}
	return nil
}

func setTitle(view string, title string) {
	v, _ := g.View(view)
	v.Title = title
}
