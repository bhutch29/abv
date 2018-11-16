package undo

import (
	"github.com/bhutch29/abv/model"
)

// CreateDrinkAction encapsulates adding a new drink to the database
type CreateDrinkAction struct {
	d model.Drink
	m model.Model
}

// NewCreateDrinkAction returns an initialized CreateDrinkAction
func NewCreateDrinkAction(d model.Drink) *CreateDrinkAction {
	c := CreateDrinkAction{}
	mod, _ := model.New()
	c.m = mod
	c.d = d
	return &c
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
