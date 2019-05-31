package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

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
		drink := model.Drink{
			Name:    trimWS(beer["beer_name"].(string)),
			Brand:   trimWS(brewery["brewery_name"].(string)),
			Abv:     beer["beer_abv"].(float64),
			Ibu:     int(beer["beer_ibu"].(float64)),
			Type:    trimWS(beer["beer_style"].(string)),
			Logo:    trimWS(brewery["brewery_label"].(string)),
			Country: trimWS(brewery["country_name"].(string)),
		}
		drinks = append(drinks, drink)
	}
	return drinks, nil
}

// trimWS trims a string of any whitespace characters defined in the Latin-1 space.
func trimWS(s string) string {
	const CutSet = " \f\t\n\r\v\x85\xA0" // TODO: also consider whitespace characters outside of the Latin-1 space
	return strings.Trim(s, CutSet)
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
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return result, err
	}

	err = validateUntappdResponse(result)
	return result, err
}

func fetchClientCredentials() (clientID, clientSecret string, err error) {
	clientID = conf.GetString("untappdId")
	if clientID == "" {
		return clientID, clientSecret, errors.New("UntappdID not supplied by client")
	}
	clientSecret = conf.GetString("untappdSecret")
	if clientSecret == "" {
		return clientID, clientSecret, errors.New("UntappdSecret not supplied by client")
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
