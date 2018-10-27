package model

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	// Registers the sqlite3 database driver
	_ "github.com/mattn/go-sqlite3"
	"os"
	"time"
)

// Model controls all the data flow into and out of the database layer
type Model struct {
	database *sqlx.DB
}

// New creates a new fully initialized Model
func New() (Model, error) {
	model := Model{}

	file := "abv.sqlite"
	if _, err := os.Stat(file); os.IsNotExist(err) {
		os.Create(file)
	}

	db, err := sqlx.Open("sqlite3", file)
	if err != nil {
		return model, err
	}

	CreateTablesIfNeeded(db)
	model.database = db
	return model, nil
}

// CreateTablesIfNeeded ensures that the database has the necessary tables
func CreateTablesIfNeeded(db *sqlx.DB) {
	db.Exec("create table if not exists Drinks (barcode integer primary key, brand varchar(255), name varchar(255), abv real, ibu real, type varchar(255), date integer)")
	db.Exec("create table if not exists Input (id integer primary key, barcode integer, quantity integer, date integer)")
	db.Exec("create table if not exists Output (id integer primary key, barcode integer, quantity integer, date integer)")
}

// Drink stores information about an available beverage
type Drink struct {
	Barcode int
	Brand   string
	Name    string
	Abv     float32
	Ibu     float32
	Type    string
	Date    int
}

// DrinkEntry defines quantities of drinks for transactions
type DrinkEntry struct {
	Barcode  int
	Quantity int
	Date     int
}

//TODO DeleteDrink
//TODO UpdateDrink

// CreateDrink adds an entry to the Drinks table, returning the id
func (m Model) CreateDrink(d Drink) (int, error) {
	now := time.Now().Unix()
	res, err := m.database.Exec(
		"insert into Drinks (barcode, brand, name, abv, ibu, type, date) Values (?, ?, ?, ?, ?, ?, ?)", d.Barcode, d.Brand, d.Name, d.Abv, d.Ibu, d.Type, now)
	if err != nil {
		return -1, err
	}
	return getID(res)
}

func (m Model) GetInventory() ([]Drink, error) {
	var drinks []Drink
	err := m.database.Select(&drinks, "select * from Drinks")
	return drinks, err
}

// InputDrinks adds an entry to the Input table, returning the id
func (m Model) InputDrinks(d DrinkEntry) (int, error) {
	now := time.Now().Unix()
	res, err := m.database.Exec(
		"insert into Input (barcode, quantity, date) Values (?, ?, ?)", d.Barcode, d.Quantity, now)
	if err != nil {
		return -1, err
	}
	return getID(res)
}

// OutputDrinks adds an entry to the Output table, returning the id
func (m Model) OutputDrinks(d DrinkEntry) (int, error) {
	now := time.Now().Unix()
	res, err := m.database.Exec(
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
