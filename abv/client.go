package main

import (
	"encoding/json"
	"fmt"
	"github.com/bhutch29/abv/model"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

// SearchUntappdByName uses the Untappd API to gather a list of Drinks that match the named search
func SearchUntappdByName(name string) ([]model.Drink, error) {
	var drinks = []model.Drink{}
	untappd, err := queryUntappdByName(name)
	if err != nil {
		return drinks, err
	}

	resp := untappd["response"].(map[string]interface{})
	beers := resp["beers"].(map[string]interface{})
	items := beers["items"].([]interface{})

	for _, item := range items {
		m := item.(map[string]interface{})
		beer := m["beer"].(map[string]interface{})
		brewery := m["brewery"].(map[string]interface{})
		var drink = model.Drink{}
		drink.Name = beer["beer_name"].(string)
		drink.Brand = brewery["brewery_name"].(string)
		drink.Abv = beer["beer_abv"].(float64)
		drink.Ibu = int(beer["beer_ibu"].(float64))
		drink.Type = beer["beer_style"].(string)
		drinks = append(drinks, drink)
	}
	return drinks, nil
}

func queryUntappdByName(name string) (map[string]interface{}, error) {
	var result map[string]interface{}
	safeName := url.QueryEscape(name)
	clientID := os.Getenv("UntappedID")
	clientSecret := os.Getenv("UntappedSecret")
	url := fmt.Sprintf("https://api.untappd.com/v4/search/beer?client_id=%s&client_secret=%s&q=%s", clientID, clientSecret, safeName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return result, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return result, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, &result)
	return result, nil
}
