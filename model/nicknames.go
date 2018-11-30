package model

import (
	"encoding/csv"
	"io"
	"log"
	"os"
)

// nickname maps formal brewery names to their shortened nicknames
var nickname = make(map[string]string)

// BrandNick returns the nickname for a given drink's brewery ("brand")
func (d Drink) BrandNick() string {
	val, ok := nickname[d.Brand]
	if ok {
		return val
	}
	return d.Brand
}

func init() {
	file, err := os.Open("nicknames.utf8")
	if err != nil {
		return
	}

	r := csv.NewReader(file)
	_, err = r.Read() // discard headers
	for {
		items, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		long := items[0]
		short := items[1]
		nickname[long] = short
	}
}
