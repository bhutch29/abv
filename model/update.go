package model

import (
	"database/sql"
	"time"
)

// ClearInputTable deletes all stocking records
func (m *Model) ClearInputTable() error {
	_, err := m.db.Exec("delete from Input")
	return err
}

// ClearOutputTable deletes all serving records
func (m *Model) ClearOutputTable() error {
	_, err := m.db.Exec("delete from Output")
	return err
}

// CreateDrink adds an entry to the Drinks table, returning the id
func (m *Model) CreateDrink(d Drink) (int, error) {
	now := time.Now().Unix()
	res, err := m.db.Exec(
	"insert into Drinks (barcode, brand, name, abv, ibu, type, shorttype, logo, country, date) Values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", d.Barcode, d.Brand, d.Name, d.Abv, d.Ibu, d.Type, d.Shorttype, d.Logo, d.Country, now)
	if err != nil {
		return -1, err
	}
	return getID(res)
}

// DeleteDrink removes an entry from the Drinks table using its barcode
func (m *Model) DeleteDrink(bc string) error {
	_, err := m.db.Exec("delete from Drinks where barcode = ?", bc)
	return err
}

// InputDrinks adds an entry to the Input table, returning the id
func (m *Model) InputDrinks(d DrinkEntry) (int, error) {
	now := time.Now().Unix()
	res, err := m.db.Exec(
		"insert into Input (barcode, quantity, date) Values (?, ?, ?)", d.Barcode, d.Quantity, now)
	if err != nil {
		return -1, err
	}
	return getID(res)
}

// UndoInputDrinks removes an entry from the Input table by id
func (m *Model) UndoInputDrinks(id int) error {
	_, err := m.db.Exec("delete from Input where id = ?", id)
	return err
}

// OutputDrinks adds an entry to the Output table, returning the id
func (m *Model) OutputDrinks(d DrinkEntry) (int, error) {
	now := time.Now().Unix()
	res, err := m.db.Exec(
		"insert into Output (barcode, quantity, date) Values (?, ?, ?)", d.Barcode, d.Quantity, now)
	if err != nil {
		return -1, err
	}
	return getID(res)
}

// UndoOutputDrinks removes an entry from the Output table by id
func (m *Model) UndoOutputDrinks(id int) error {
	_, err := m.db.Exec("delete from Output where id = ?", id)
	return err
}

func getID(result sql.Result) (int, error) {
	id, err := result.LastInsertId()
	if err != nil {
		return -1, err
	}
	return int(id), nil
}
