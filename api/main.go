package main

import (
	"encoding/json"
	"github.com/bhutch29/abv/model"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
	"io"
	"log"
	"net/http"
	"flag"
	"fmt"
	"os"
	"path"
)

var (
	m model.Model
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

	corsEnabledHandler := cors.Default().Handler(router)
	log.Fatal(http.ListenAndServe(":8081", corsEnabledHandler))
}

func handleFlags(){
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
	setHeader(w)
	drinks, err := m.GetInventory()
	cacheImages(drinks)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(drinks)
}

func cacheImages(drinks []model.StockedDrink) {
	for _, drink := range drinks {
		cacheImageFromURL(drink.Logo)
	}
}

func cacheImageFromURL(url string) {
	file := path.Base(url)
	if !exists("images/" + file) {
		response, err := http.Get(url)
		if err != nil {
			return
		}
		defer response.Body.Close()

		_ = os.MkdirAll("images", os.ModePerm)
		f, err := os.Create("images/" + file)
		if err != nil {
			log.Fatal("Could not create file")
		}
		defer f.Close()

		_, _ = io.Copy(f, response.Body)
	}
}

func exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func setHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}
