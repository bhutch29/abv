package model

import "database/sql"
import "strings"

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

// GetAllStoredDrinks returns every saved Drink row in the database
func (m *Model) GetAllStoredDrinks() ([]Drink, error) {
	var drinks []Drink
	err := m.db.Select(&drinks, "select * from Drinks")
	drinks = m.setDrinksNicknames(drinks)
	return drinks, err
}

// TODO: These DrinkNickname methods are pretty awful, lots of repeated code. How to improve?
func (m *Model) setDrinksNicknames(drinks []Drink) []Drink {
	brandNicks := m.conf.GetStringMapString("breweryNicknames")
	beerNicks := m.conf.GetStringMapString("beerNicknames")
	styleNicks := m.conf.GetStringMapString("styleNicknames")
	var result []Drink
	for _, drink := range drinks {
		if nick, ok := brandNicks[strings.ToLower(drink.Brand)]; ok {
			drink.Brand = nick
		}
		if nick, ok := beerNicks[strings.ToLower(drink.Name)]; ok {
			drink.Name = nick
		}
		if nick, ok := styleNicks[strings.ToLower(drink.Type)]; ok {
			drink.Type = nick
			drink.Shorttype = nick
		}
		result = append(result, drink)
	}
	return result
}

func (m *Model) setStockedDrinksNicknames(drinks []StockedDrink) []StockedDrink {
	brandNicks := m.conf.GetStringMapString("breweryNicknames")
	beerNicks := m.conf.GetStringMapString("beerNicknames")
	styleNicks := m.conf.GetStringMapString("styleNicknames")
	var result []StockedDrink
	for _, drink := range drinks {
		if nick, ok := brandNicks[strings.ToLower(drink.Brand)]; ok {
			drink.Brand = nick
		}
		if nick, ok := beerNicks[strings.ToLower(drink.Name)]; ok {
			drink.Name = nick
		}
		if nick, ok := styleNicks[strings.ToLower(drink.Type)]; ok {
			drink.Type = nick
			drink.Shorttype = nick
		}
		result = append(result, drink)
	}
	return result
}

func (m *Model) setDrinkNickname(drink Drink) Drink {
	brandNicks := m.conf.GetStringMapString("breweryNicknames")
	beerNicks := m.conf.GetStringMapString("beerNicknames")
	styleNicks := m.conf.GetStringMapString("styleNicknames")
	if nick, ok := brandNicks[strings.ToLower(drink.Brand)]; ok {
		drink.Brand = nick
	}
	if nick, ok := beerNicks[strings.ToLower(drink.Name)]; ok {
		drink.Name = nick
	}
	if nick, ok := styleNicks[strings.ToLower(drink.Type)]; ok {
		drink.Type = nick
		drink.Shorttype = nick
	}
	return drink
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
	d = m.setDrinkNickname(d)
	return d, err
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
  end as quantity
from Drinks as A

left join (
  select barcode, sum(quantity) as InputQuantity
  from Input
  group by barcode
) as B
on A.Barcode = B.Barcode

left join (
  select barcode, sum(quantity) as OutputQuantity
  from Output
  group by barcode
) as C
on A.Barcode = C.Barcode

where quantity > 0
order by A.Brand
`
	err := m.db.Select(&result, sql)
	result = m.setStockedDrinksNicknames(result)
	return result, err
}

// GetInputWithinDateRange returns every drink inputted within a date range, inclusive
func (m *Model) GetInputWithinDateRange(dates DateRange) (result []StockedDrink, err error) {
	sql := `
select A.*,
  case
    when C.InputQuantity is null then 0
    else C.InputQuantity
  end as quantity
from Drinks as A

left join (
  select barcode, sum(quantity) as InputQuantity
  from Input as O where O.Date >= ? and O.Date <= ?
  group by barcode
) as C
on A.Barcode = C.Barcode

where quantity > 0
order by A.Brand
`
	err = m.db.Select(&result, sql, dates.Start, dates.End)
	return result, err
}

// GetOutputWithinDateRange returns every drink served within a date range, inclusive
func (m *Model) GetOutputWithinDateRange(dates DateRange) (result []StockedDrink, err error) {
	sql := `
select A.*,
  case
    when C.OutputQuantity is null then 0
    else C.OutputQuantity
  end as quantity
from Drinks as A

left join (
  select barcode, sum(quantity) as OutputQuantity
  from Output as O where O.Date >= ? and O.Date <= ?
  group by barcode
) as C
on A.Barcode = C.Barcode

where quantity > 0
order by A.Brand
`
	err = m.db.Select(&result, sql, dates.Start, dates.End)
	return result, err
}
