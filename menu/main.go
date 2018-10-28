package main

import (
	"github.com/bhutch29/abv/model"
	"fmt"
	"net/http"
	"log"
	"encoding/json"
	"time"
)

func main() {
	url := "http://localhost:8081/inventory"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("NewRequest: ", err)
		return
	}

	client := &http.Client{}

	for {
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal("Do: ", err)
			return
		}

		defer resp.Body.Close()
		var inventory []model.Drink
		if err := json.NewDecoder(resp.Body).Decode(&inventory); err != nil {
			log.Println(err)
		}
		fmt.Println()
		fmt.Println(inventory)
		time.Sleep(2 * time.Second)
	}
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
