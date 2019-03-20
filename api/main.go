package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"net/url"

	"github.com/bhutch29/abv/cache"
	"github.com/bhutch29/abv/model"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
	"golang.org/x/text/unicode/norm"
)

var (
	m       model.Model
	version = "undefined"
)

func main() {
	handleFlags()

	mod, err := model.New()
	if err != nil {
		log.Fatal(err)
	}
	m = mod

	router := httprouter.New()

	router.GET("/health", healthCheck)
	router.GET("/inventory", getInventory)
	router.GET("/inventory/quantity", getInventoryQuantity)
	router.GET("/inventory/variety", getInventoryVariety)
	router.GET("/inventory/sorted/:sortFields", getInventorySorted)

	corsEnabledHandler := cors.Default().Handler(router)
	log.Fatal(http.ListenAndServe(":8081", corsEnabledHandler))
}

func handleFlags() {
	ver := flag.Bool("version", false, "Prints the version")
	flag.Parse()

	if *ver {
		fmt.Println(version)
		os.Exit(0)
	}
}

func healthCheck(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.WriteHeader(http.StatusOK)
	setHeader(w)
	// TODO: Expand Health Check
	io.WriteString(w, `{"alive": true}`)
}

func getInventory(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	drinks, err := m.GetInventory()
	drinks = abbreviate(drinks)
	encodeDrinks(drinks, err, w)
}

func abbreviate(drinksIn []model.StockedDrink) (drinksOut []model.StockedDrink) {
	const DrinkLenMax = 28
	for _, drink := range drinksIn {
		nfcBytes := norm.NFC.Bytes([]byte(drink.Name))
		nfcRunes := []rune(string(nfcBytes))
		visualLen := len(nfcRunes)
		if visualLen <= DrinkLenMax {
			drinksOut = append(drinksOut, drink)
			continue
		}
		drink.Drink.Name = string(nfcRunes[:DrinkLenMax])
		drinksOut = append(drinksOut, drink)
	}
	return drinksOut
}

func getInventoryQuantity(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	q, err := m.GetInventoryTotalQuantity()
	encodeValue(q, err, w)
}

func getInventoryVariety(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	q, err := m.GetInventoryTotalVariety()
	encodeValue(q, err, w)
}

func getInventorySorted(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	res, err := url.ParseQuery(ps.ByName("sortFields"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	drinks, err := m.GetInventorySorted(res["sortBy"])
	encodeDrinks(drinks, err, w)
}

func encodeValue(val interface{}, err error, w http.ResponseWriter) {
	setHeader(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(val)
}

func encodeDrinks(drinks []model.StockedDrink, err error, w http.ResponseWriter) {
	setHeader(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	for _, drink := range drinks {
		err = cache.Image(drink.Logo)
		if err != nil {
			log.Println("Failed HTTP request while caching image for drink: ", drink.Brand, " ", drink.Name)
		}
	}

	json.NewEncoder(w).Encode(drinks)
}

func setHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}
