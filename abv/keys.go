package main

import (
	"fmt"
	"github.com/jroimartin/gocui"
	aur "github.com/logrusorgru/aurora"
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
		if k.viewname == "" {
			result = result + fmt.Sprintf("%s->%s ", aur.Green(k.shortkey), k.shortname)
		}
	}
	return result
}

func configureKeys() error {
	for _, key := range keys {
		if err := g.SetKeybinding(key.viewname, key.key, gocui.ModNone, key.handler); err != nil {
			return err
		}
	}

	return nil
}
