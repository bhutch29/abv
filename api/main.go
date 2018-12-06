package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/bhutch29/abv/cache"
	"github.com/bhutch29/abv/model"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
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
	router.GET("/inventory/sorted/:sortField", getInventorySorted)

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
	encodeDrinks(drinks, err, w)
}

func getInventorySorted(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	drinks, err := m.GetInventorySorted(ps.ByName("sortField"))
	encodeDrinks(drinks, err, w)
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
