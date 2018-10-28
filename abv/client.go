package main

import (
	"encoding/json"
	"net/http"
	"fmt"
	"net/url"
	"os"
	"github.com/bhutch29/abv/model"
)

// SearchUntappdByName uses the Untappd API to gather a list of Drinks that match the named search
func SearchUntappdByName(name string) ([]model.Drink, error) {
	var drinks = []model.Drink{}
	untappd, err := queryUntappdByName(name)
	if err != nil {
		return drinks, err
	}
	for _, item := range untappd.Response.Beers.Items {
		var drink = model.Drink{}
		drink.Name = item.Beer.BeerName
		drink.Brand = item.Brewery.BreweryName
		drink.Abv = item.Beer.BeerAbv
		drink.Ibu = item.Beer.BeerIbu
		drink.Type = item.Beer.BeerStyle
		drinks = append(drinks, drink)
	}
	logToMain(drinks)
	return drinks, nil
}

func queryUntappdByName(name string) (untappdSearch, error) {
	var results = untappdSearch{}
	safeName := url.QueryEscape(name)
	clientID := os.Getenv("UntappdId")
	clientSecret := os.Getenv("UntappdSecret")
	url := fmt.Sprintf("https://api.untappd.com/v4/search/beer?client_id=%s&client_secret=%s&q=%s", clientID, clientSecret, safeName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return results, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return results, err
	}

	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return results, err
	}
	return results, nil
}

type untappdSearch struct {
	Meta struct {
		Code         int `json:"code"`
		ResponseTime struct {
			Time    float64 `json:"time"`
			Measure string  `json:"measure"`
		} `json:"response_time"`
		InitTime struct {
			Time    int    `json:"time"`
			Measure string `json:"measure"`
		} `json:"init_time"`
	} `json:"meta"`
	Notifications []interface{} `json:"notifications"`
	Response      struct {
		Message       string  `json:"message"`
		TimeTaken     float64 `json:"time_taken"`
		BreweryID     int     `json:"brewery_id"`
		SearchType    string  `json:"search_type"`
		TypeID        int     `json:"type_id"`
		SearchVersion int     `json:"search_version"`
		Found         int     `json:"found"`
		Offset        int     `json:"offset"`
		Limit         int     `json:"limit"`
		Term          string  `json:"term"`
		ParsedTerm    string  `json:"parsed_term"`
		Beers         struct {
			Count int `json:"count"`
			Items []struct {
				CheckinCount int  `json:"checkin_count"`
				HaveHad      bool `json:"have_had"`
				YourCount    int  `json:"your_count"`
				Beer         struct {
					Bid             int     `json:"bid"`
					BeerName        string  `json:"beer_name"`
					BeerLabel       string  `json:"beer_label"`
					BeerAbv         float64 `json:"beer_abv"`
					BeerSlug        string  `json:"beer_slug"`
					BeerIbu         int     `json:"beer_ibu"`
					BeerDescription string  `json:"beer_description"`
					CreatedAt       string  `json:"created_at"`
					BeerStyle       string  `json:"beer_style"`
					InProduction    int     `json:"in_production"`
					AuthRating      int     `json:"auth_rating"`
					WishList        bool    `json:"wish_list"`
				} `json:"beer"`
				Brewery struct {
					BreweryID      int    `json:"brewery_id"`
					BreweryName    string `json:"brewery_name"`
					BrewerySlug    string `json:"brewery_slug"`
					BreweryPageURL string `json:"brewery_page_url"`
					BreweryType    string `json:"brewery_type"`
					BreweryLabel   string `json:"brewery_label"`
					CountryName    string `json:"country_name"`
					Contact        struct {
						Twitter   string `json:"twitter"`
						Facebook  string `json:"facebook"`
						Instagram string `json:"instagram"`
						URL       string `json:"url"`
					} `json:"contact"`
					Location struct {
						BreweryCity  string  `json:"brewery_city"`
						BreweryState string  `json:"brewery_state"`
						Lat          float64 `json:"lat"`
						Lng          float64 `json:"lng"`
					} `json:"location"`
					BreweryActive int `json:"brewery_active"`
				} `json:"brewery"`
			} `json:"items"`
		} `json:"beers"`
		Homebrew struct {
			Count int `json:"count"`
			Items []struct {
				CheckinCount int  `json:"checkin_count"`
				HaveHad      bool `json:"have_had"`
				YourCount    int  `json:"your_count"`
				Beer         struct {
					Bid             int     `json:"bid"`
					BeerName        string  `json:"beer_name"`
					BeerLabel       string  `json:"beer_label"`
					BeerAbv         float64 `json:"beer_abv"`
					BeerSlug        string  `json:"beer_slug"`
					BeerIbu         int     `json:"beer_ibu"`
					BeerDescription string  `json:"beer_description"`
					CreatedAt       string  `json:"created_at"`
					BeerStyle       string  `json:"beer_style"`
					InProduction    int     `json:"in_production"`
					AuthRating      int     `json:"auth_rating"`
					WishList        bool    `json:"wish_list"`
				} `json:"beer"`
				Brewery struct {
					BreweryID      int    `json:"brewery_id"`
					BreweryName    string `json:"brewery_name"`
					BrewerySlug    string `json:"brewery_slug"`
					BreweryPageURL string `json:"brewery_page_url"`
					BreweryType    string `json:"brewery_type"`
					BreweryLabel   string `json:"brewery_label"`
					CountryName    string `json:"country_name"`
					Contact        struct {
						Twitter   string `json:"twitter"`
						Facebook  string `json:"facebook"`
						Instagram string `json:"instagram"`
						URL       string `json:"url"`
					} `json:"contact"`
					Location struct {
						BreweryCity  string `json:"brewery_city"`
						BreweryState string `json:"brewery_state"`
						Lat          int    `json:"lat"`
						Lng          int    `json:"lng"`
					} `json:"location"`
					BreweryActive int `json:"brewery_active"`
				} `json:"brewery"`
			} `json:"items"`
		} `json:"homebrew"`
		Breweries struct {
			Items []interface{} `json:"items"`
			Count int           `json:"count"`
		} `json:"breweries"`
	} `json:"response"`
}
