package model

import (
	"database/sql"
	"time"
)

// DeleteDrink removes an entry from the Drinks table using its barcode
func (m *Model) DeleteDrink(bc string) error {
	_, err := m.db.Exec("delete from Drinks where barcode = ?", bc)
	return err
}

// CreateDrink adds an entry to the Drinks table, returning the id
func (m *Model) CreateDrink(d Drink) (int, error) {
	now := time.Now().Unix()
	res, err := m.db.Exec(
		"insert into Drinks (barcode, brand, name, abv, ibu, type, logo, date) Values (?, ?, ?, ?, ?, ?, ?, ?)", d.Barcode, d.Brand, d.Name, d.Abv, d.Ibu, d.Type, d.Logo, now)
	if err != nil {
		return -1, err
	}
	return getID(res)
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

func getID(result sql.Result) (int, error) {
	id, err := result.LastInsertId()
	if err != nil {
		return -1, err
	}
	return int(id), nil
}
