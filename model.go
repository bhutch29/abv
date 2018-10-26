package model

import (
	"database/sql"
	// TODO: Can't remember why a blank was needed here, but it fixed a bug.
	_ "github.com/mattn/go-sqlite3"
	"github.com/jmoiron/sqlx"
)

//create table Drinks
//(
//barcode       integer primary key,
//brand         varchar(255),
//name          varchar(255),
//abv           integer,
//type          varchar(255),
//date          integer
//)

//create table Input
//(
//id            integer primary key,
//barcode       integer,
//quantity      integer,
//date          integer
//)

//create table Output
//(
//id            integer primary key,
//barcode       integer,
//quantity      integer,
//date          integer
//)

// Model controls all the data flow into and out of the database layer
type Model struct{
	database *sqlx.DB
}

// New creates a new fully initialized Model
func New() (Model, error) {
	model := Model{}
	db, err := sqlx.Open("sqlite3", "./abv.sqlite")
	if err != nil {
		return model, err
	}
	model.database = db
	return model, nil
}

// Drink stores information about an available beverage
type Drink struct{
	Barcode int
	Brand string
	Name string
	Abv int
	Type int
	Date int
}

// DrinkEntry defines quantities of drinks for transactions
type DrinkEntry struct{
	Barcode int
	Quantity int
	Date int
}

// CreateDrink adds an entry to the Drinks table, returning the id
func (m Model) CreateDrink(d Drink) (int, error){
	now := time.Now().Unix()
	res, err := m.database.Exec(
	"insert into Drinks (barcode, brand, name, abv, type, date) Values (?, ?, ?, ?, ?, ?)", d.Barcode, d.Brand, d.Name, d.Abv, d.Type, now)
}

// InputDrinks adds an entry to the Input table, returning the id
func (m Model) InputDrinks(d DrinkEntry) (int, error){
	now := time.Now().Unix()
	res, err := m.database.Exec(
		"insert into Input (barcode, quantity, date) Values (?, ?, ?)", d.Barcode, d.Quantity, now)
	if err != nil {
		return -1, err
	}
	return getID(res)
}

// OutputDrinks adds an entry to the Output table, returning the id
func (m Model) OutputDrinks(d DrinkEntry) (int, error){
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
