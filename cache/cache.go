package cache

import (
	"path"
	"net/http"
	"os"
	"log"
	"io"
)

// Image queries and saves an image from the provided url if it isn't cached already
func Image(url string) {
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
