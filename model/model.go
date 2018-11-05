package model

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	// Registers the sqlite3 db driver
	_ "github.com/mattn/go-sqlite3"
	"os"
	"time"
)

// Model controls all the data flow into and out of the db layer
type Model struct {
	db *sqlx.DB
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

	model.db = db
	model.CreateTablesIfNeeded()
	return model, nil
}

// CreateTablesIfNeeded ensures that the db has the necessary tables
func (m *Model) CreateTablesIfNeeded() {
	m.db.Exec("create table if not exists Drinks (barcode varchar(255) primary key, brand varchar(255), name varchar(255), abv real, ibu real, type varchar(255), date integer)")
	m.db.Exec("create table if not exists Input (id integer primary key, barcode integer, quantity integer, date integer)")
	m.db.Exec("create table if not exists Output (id integer primary key, barcode integer, quantity integer, date integer)")
}

// Drink stores information about an available beverage
type Drink struct {
	Barcode string
	Brand   string
	Name    string
	Abv     float64
	Ibu     int
	Type    string
	Date    int64
}

// DrinkEntry defines quantities of drinks for transactions
type DrinkEntry struct {
	Barcode  string
	Quantity int
	Date     int64
}

// StockedDrink is an extension of drink with an additional field for quantity
type StockedDrink struct {
	Drink
	Quantity int
}

// GetInventory returns every drink with at least one quantity in stock
func (m *Model) GetInventory() ([]StockedDrink, error) {
	var result []StockedDrink
	sql := `
select A.*,
case
  when B.InputQuantity is null then 0
  when C.OutputQuantity is null then B.InputQuantity
  else (B.InputQuantity - C.OutputQuantity)
end quantity
from Drinks A
left join (
  select barcode, sum(quantity) as InputQuantity
  from Input
  group by barcode
)
B
on A.Barcode = B.Barcode
left join (
  select barcode, sum(quantity) as OutputQuantity
  from Output
  group by barcode
)
C
on A.Barcode = C.Barcode
where quantity > 0
order by A.Brand
`
	err := m.db.Select(&result, sql)
	return result, err
}

// GetCountByBarcode returns the total number of currently stocked beers with a specific barcode
func (m *Model) GetCountByBarcode(bc string) (int, error) {
	var input, output int
	if err := m.db.Get(&input, "select case when sum(quantity)is null then 0 else sum(quantity) end quantity from Input where barcode = ?", bc); err != nil {
		return -1, err
	}
	if err := m.db.Get(&output, "select case when sum(quantity) is null then 0 else sum(quantity) end quantity from Output where barcode = ?", bc); err != nil {
		return -1, err
	}

	return input - output, nil
}

// GetDrinkByBarcode returns all stored information about a drink based on its barcode
func (m *Model) GetDrinkByBarcode(bc string) (Drink, error) {
	var d Drink
	err := m.db.Get(&d, "select * from Drinks where barcode = ?", bc)
	//TODO Check that a value got returned, or at least throws a sql.Err___ if nothing found
	return d, err
}

// BarcodeExists checks if a barcode is already in the database
func (m *Model) BarcodeExists(bc string) (bool, error) {
	var barcode string
	err := m.db.Get(&barcode, "select barcode from Drinks where barcode = ? limit 1", bc)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	if barcode == bc {
		return true, nil
	}
	return false, nil
}

// DeleteDrink removes an entry from the Drinks table using its barcode
func (m *Model) DeleteDrink(bc string) error {
	_, err := m.db.Exec("delete from Drinks where barcode = ?", bc)
	return err
}

// CreateDrink adds an entry to the Drinks table, returning the id
func (m *Model) CreateDrink(d Drink) (int, error) {
	now := time.Now().Unix()
	res, err := m.db.Exec(
		"insert into Drinks (barcode, brand, name, abv, ibu, type, date) Values (?, ?, ?, ?, ?, ?, ?)", d.Barcode, d.Brand, d.Name, d.Abv, d.Ibu, d.Type, now)
	if err != nil {
		return -1, err
	}
	return getID(res)
}

// GetAllStoredDrinks returns every saved Drink row in the database
func (m *Model) GetAllStoredDrinks() ([]Drink, error) {
	var drinks []Drink
	err := m.db.Select(&drinks, "select * from Drinks")
	return drinks, err
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
