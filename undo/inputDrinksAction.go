package undo

import (
	"github.com/bhutch29/abv/model"
)

// InputDrinksAction encapsulates stocking a drink
type InputDrinksAction struct {
	id int
	de model.DrinkEntry
	m  model.Model
}

// NewInputDrinksAction returns an initialized InputDrinksAction
func NewInputDrinksAction(de model.DrinkEntry) *InputDrinksAction {
	i := InputDrinksAction{}
	mod, _ := model.New()
	i.m = mod
	i.de = de
	return &i
}

// Do implements the ReversibleAction interface
func (a *InputDrinksAction) Do() error {
	i, err := a.m.InputDrinks(a.de)
	if err != nil {
		return err
	}
	a.id = i
	return nil
}

// Undo implements the ReversibleAction interface
func (a *InputDrinksAction) Undo() error {
	err := a.m.UndoInputDrinks(a.id)
	return err
}
