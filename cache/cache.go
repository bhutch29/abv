package cache

import (
	"path"
	"net/http"
	"os"
	"log"
	"io"
)

// Images queries and saves images from the provided urls if they aren't cached already
func Images(urls []string) {
	for _, url := range urls {
		cacheImageFromURL(url)
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
