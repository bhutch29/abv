package undo

import (
	"github.com/bhutch29/abv/model"
)

// CreateAndInputAction encapsulates adding a new drink to the database and inputting it
type CreateAndInputAction struct {
	c CreateDrinkAction
	i InputDrinksAction
}

// NewCreateAndInputAction returns an initialized CreateAndInputAction
func NewCreateAndInputAction(d model.Drink, de model.DrinkEntry) CreateAndInputAction {
	a := CreateAndInputAction{}
	c := NewCreateDrinkAction(d)
	i := NewInputDrinksAction(de)
	a.c = c
	a.i = i
	return a
}

// Do implements the ReversibleAction interface
func (a *CreateAndInputAction) Do() error {
	var err error
	err = a.c.Do()
	err = a.i.Do()
	return err
}

// Undo implements the ReversibleAction interface
func (a *CreateAndInputAction) Undo() error {
	var err error
	err = a.i.Undo()
	err = a.c.Undo()
	return err
}
