package undo

import (
	"github.com/bhutch29/abv/model"
)

// OutputDrinksAction encapsulates serving a drink
type OutputDrinksAction struct {
	id int
	de model.DrinkEntry
	m  model.Model
}

// NewOutputDrinksAction returns an initialized OutputDrinksAction
func NewOutputDrinksAction(de model.DrinkEntry) *OutputDrinksAction {
	o := OutputDrinksAction{}
	mod, _ := model.New()
	o.m = mod
	o.de = de
	return &o
}

// Do implements the ReversibleAction interface
func (a *OutputDrinksAction) Do() error {
	i, err := a.m.OutputDrinks(a.de)
	if err != nil {
		return err
	}
	a.id = i
	return nil
}

// Undo implements the ReversibleAction interface
func (a *OutputDrinksAction) Undo() error {
	err := a.m.UndoOutputDrinks(a.id)
	return err
}
