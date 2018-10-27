package main

import (
	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
	"io"
	"log"
	"net/http"
	"encoding/json"
	"github.com/bhutch29/abv/model"
)

var m model.Model

func main() {
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

func healthCheck(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.WriteHeader(http.StatusOK)
	setHeader(w)

	// TODO: Expand Health Check
	io.WriteString(w, `{"alive": true}`)
}

func getInventory(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	setHeader(w)
	drinks, err := m.GetInventory()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(drinks)
}

func setHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}
