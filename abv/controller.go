package main

import (
	"github.com/bhutch29/abv/model"
)

// Mode is an Enum of operating modes
type Mode int

const (
	serving Mode = iota
	stocking
	administration
)

// Controller defines methods required to communicate with the ABV backend
type Controller interface {
	GetMode() Mode
	SetMode(Mode)
	CreateDrink(model.Drink)
	HandleScannedDrink(model.Drink)
}

// ModalController supports using the GUI via distinct behavioral modes
type ModalController struct {
	currentMode Mode
	backend     model.Model
}

// New creates a new fully initialized ModalController
func New() (Controller, error) {
	m := &ModalController{}
	m.currentMode = serving
	backend, err := model.New()
	if err != nil {
		m.backend = backend
		return m, err
	}
	return m, nil
}

// GetMode returns the current operating Mode
func (c *ModalController) GetMode() Mode {
	return c.currentMode
}

// SetMode changes the current operating Mode
func (c *ModalController) SetMode(m Mode) {
	c.currentMode = m
}

// CreateDrink stores a new Drink in the scanning database
func (c *ModalController) CreateDrink(d model.Drink) {
	//TODO
}

// HandleScannedDrink processes a drink after it has been scanned. Behavior varies based on operating Mode
func (c *ModalController) HandleScannedDrink(d model.Drink) {
	//TODO
}
