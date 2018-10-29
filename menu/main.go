package main

import (
	"encoding/json"
	"fmt"
	"github.com/bhutch29/abv/model"
	"log"
	"net/http"
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
