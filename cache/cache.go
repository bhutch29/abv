package cache

import (
	"path"
	"net/http"
	"os"
	"log"
	"io"
	"github.com/bhutch29/abv/config"
)

// Image queries and saves an image from the provided url if it isn't cached already
func Image(url string) {
	conf, err := config.New()
	if err != nil {
		log.Fatal("Could not get configuration: ", err)
	}
	imagePath := conf.GetString("imageCachePath")

	file := path.Base(url)
	if !exists(imagePath + "/" + file) {
		response, err := http.Get(url)
		if err != nil {
			return
		}

		_ = os.MkdirAll(imagePath, os.ModePerm)
		f, err := os.Create(imagePath + "/" + file)
		if err != nil {
			log.Fatal("Could not create file: ", err)
		}

		_, _ = io.Copy(f, response.Body)

		response.Body.Close()
		f.Close()
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
