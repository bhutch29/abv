package undo

import (
	"github.com/bhutch29/abv/model"
)

// ReversibleAction defines an undo-able database change
type ReversibleAction interface {
	Do() error
	Undo() error
}

// CreateDrinkAction encapsulates adding a new drink to the database
type CreateDrinkAction struct {
	d model.Drink
	m model.Model
}

// Do implements the ReversibleAction interface
func (a *CreateDrinkAction) Do() error {
	_, err := a.m.CreateDrink(a.d)
	return err
}

// Undo implements the ReversibleAction interface
func (a *CreateDrinkAction) Undo() error {
	err := a.m.DeleteDrink(a.d.Barcode)
	return err
}

// InputDrinksAction encapsulates stocking a drink
type InputDrinksAction struct {
	id int
	de model.DrinkEntry
	m model.Model
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

// OutputDrinksAction encapsulates serving a drink
type OutputDrinksAction struct {
	id int
	de model.DrinkEntry
	m model.Model
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
