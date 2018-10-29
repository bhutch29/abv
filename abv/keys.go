package main

import (
	"fmt"

	"github.com/jroimartin/gocui"
)

type key struct {
	viewname  string
	key       interface{}
	handler   func(*gocui.Gui, *gocui.View) error
	shortkey  string
	shortname string
}

func generateKeybindString() string {
	var result string
	for _, k := range keys {
		s := fmt.Sprintf("%s->%s ", k.shortkey, k.shortname)
		result = result + s
	}
	return result
}

func configureKeys(g *gocui.Gui) error {
	for _, key := range keys {
		if err := g.SetKeybinding(key.viewname, key.key, gocui.ModNone, key.handler); err != nil {
			return err
		}
	}

	return nil
}
