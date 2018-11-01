package main

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/sirupsen/logrus"
)

var popupDisplayed = false

// Define Prompt dimensions
const (
	inputHeight    = 4
	inputCursorPos = 4
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
		logFile.Info(v.Name())
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

		v.Title = "Choose desired drink..."
		v.Frame = true
		v.Highlight = true
		//TODO: Set Selected FG and BG colors
		g.SetViewOnBottom(popup)
	}
	return nil
}

func togglePopup(){
	vn := popup

	if !popupDisplayed {
		g.SetViewOnTop(vn)
		g.SetCurrentView(vn)
	} else {
		g.SetViewOnBottom(vn)
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

func clearInput() {
	g.Update(func(g *gocui.Gui) error {
		input, err := g.View(input)
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

func hideError() {
	g.DeleteView(errorView)
}

func popupScrollUp(g *gocui.Gui, v *gocui.View) error {
	err := moveViewCursorUp(v, 0)
	if err != nil {
		logFile.Error(err)
		logGui.Error(err)
	}
	return err //TODO: Verify what happens if error is returned to gocui layer like this. Should we return nil?
}

func popupScrollDown(g *gocui.Gui, v *gocui.View) error {
	err := moveViewCursorDown(v, false)
	if err != nil {
		logFile.Error(err)
		logGui.Error(err)
	}
	return err
}

func popupSelectItem(g *gocui.Gui, v *gocui.View) error {
	line, err := getViewLine(v)
	togglePopup()
	resetViewCursor(v)
	logFile.WithFields(logrus.Fields{
		"category": "userEntry",
		"entry": line,
	}).Info("User selected a beer")

	//TODO Do something with selected value
	logGui.Info("You selected: " + line)
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

func moveViewCursorUp(v *gocui.View, dY int) error {
	ox, oy := v.Origin()
	cx, cy := v.Cursor()
	if cy > dY {
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
