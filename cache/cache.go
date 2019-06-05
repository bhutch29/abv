// Package cache provides methods for caching brand logo images.
package cache

import (
	"io"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/bhutch29/abv/config"
	"github.com/spf13/viper"
)

var conf *viper.Viper

func init() {
	var err error
	conf, err = config.New()
	if err != nil {
		log.Fatal("Could not get configuration: ", err)
	}
}

// Image queries and saves an image from the provided url if it isn't cached already
func Image(url string) error {
	imagePath := path.Join(conf.GetString("configPath"), "images")
	file := path.Join(imagePath, path.Base(url))
	if exists(file) {
		return nil
	}
	response, err := http.Get(url)
	if err != nil {
		return err
	}

	_ = os.MkdirAll(imagePath, os.ModePerm)
	f, err := os.Create(file)
	if err != nil {
		log.Fatal("Could not create file: ", err)
	}

	if _, err = io.Copy(f, response.Body); err != nil {
		return err
	}

	response.Body.Close()
	f.Close()
	return nil
}

// exists returns whether or not a file or directory exists.
func exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
