package model

import (
	"os"

	"github.com/jmoiron/sqlx"
	// Registers the sqlite3 db driver
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
	"github.com/bhutch29/abv/config"
	"github.com/mitchellh/go-homedir"
)

// Model controls all the data flow into and out of the db layer
type Model struct {
	db *sqlx.DB
	conf *viper.Viper
}

// New creates a new fully initialized Model
func New() (Model, error) {
	model := Model{}
	conf, err := config.New()
	if err != nil {
		return model, err
	}
	model.conf = conf

	configPath, _ := homedir.Expand((conf.GetString("configPath")))
	file := configPath + "/abv.sqlite"
	if _, err := os.Stat(file); os.IsNotExist(err) {
		os.Create(file)
	}

	db, err := sqlx.Open("sqlite3", file)
	if err != nil {
		return model, err
	}

	model.db = db
	model.CreateTablesIfNeeded()
	return model, nil
}

// CreateTablesIfNeeded ensures that the db has the necessary tables
func (m *Model) CreateTablesIfNeeded() {
	m.db.Exec(`
create table if not exists Drinks (
barcode varchar(255) primary key,
brand varchar(255),
name varchar(255),
abv real,
ibu real,
type varchar(255),
shorttype varchar(255),
logo varchar(255),
country varchar(255),
date integer)
`)
	m.db.Exec(`
create table if not exists Input (
id integer primary key,
barcode varchar(255),
quantity integer,
date integer)
`)
	m.db.Exec(`
create table if not exists Output (
id integer primary key,
barcode varchar(255),
quantity integer,
date integer)
`)
}

// Date is a representation of a Unix time stamp
type Date int64

// DateRange is an inclusive range of dates
type DateRange struct {
	Start Date
	End   Date
}

// Drink stores information about an available beverage
type Drink struct {
	Barcode   string
	Brand     string
	Name      string
	Abv       float64
	Ibu       int
	Type      string
	Shorttype string
	Logo      string
	Date      Date
	Country   string
}

// DrinkEntry defines quantities of drinks for transactions
type DrinkEntry struct {
	Barcode  string
	Quantity int
	Date     Date
}

// StockedDrink is an extension of drink with an additional field for quantity
type StockedDrink struct {
	Drink
	Quantity int
}
