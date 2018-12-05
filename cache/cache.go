package cache

import (
	"io"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/bhutch29/abv/config"
)

// Image queries and saves an image from the provided url if it isn't cached already
func Image(url string) error {
	conf, err := config.New()
	if err != nil {
		log.Fatal("Could not get configuration: ", err)
	}
	imagePath := conf.GetString("imageCachePath")

	file := path.Join(imagePath, path.Base(url))
	if !exists(file) {
		response, err := http.Get(url)
		if err != nil {
			return err
		}

		_ = os.MkdirAll(imagePath, os.ModePerm)
		f, err := os.Create(file)
		if err != nil {
			log.Fatal("Could not create file: ", err)
		}

		_, _ = io.Copy(f, response.Body)

		response.Body.Close()
		f.Close()
	}
	return nil
}

func exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
