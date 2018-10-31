package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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
	clientID, clientSecret, err := fetchClientCredentials()
	if err != nil {
		return result, err
	}

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
	err = validateUntappdResponse(result)
	if err != nil {
		return result, err
	}
	return result, nil
}

func fetchClientCredentials() (clientID, clientSecret string, err error) {
	clientID = os.Getenv("UntappdID")
	if clientID == "" {
		return clientID, clientSecret, fmt.Errorf("UntappdID not supplied by client")
	}
	clientSecret = os.Getenv("UntappdSecret")
	if clientSecret == "" {
		return clientID, clientSecret, fmt.Errorf("UntappdSecret not supplied by client")
	}
	return clientID, clientSecret, nil
}

func validateUntappdResponse(response map[string]interface{}) (err error) {
	meta := response["meta"].(map[string]interface{})
	code := int(meta["code"].(float64))
	if code != http.StatusOK {
		return fmt.Errorf("Untappd status code %v: %v", code, http.StatusText(code))
	}
	return nil
}
