package main

import (
	"encoding/json"
	"github.com/bhutch29/abv/model"
	"github.com/julienschmidt/httprouter"
	"io/ioutil"
	"log"
	"net/http"
	"time"
	"flag"
	"fmt"
	"os"
	"github.com/bhutch29/abv/config"
	"github.com/spf13/viper"
	"path"
)

var (
	conf *viper.Viper
	version = "undefined"
)

// Page is the backing type for all pages
type Page struct {
}

// FrontPage is the primary view object
type FrontPage struct {
	Page
	Drinks []model.StockedDrink
}

func stringFromFile(path string) string {
	b, _ := ioutil.ReadFile(path)
	return string(b)
}

func newFrontPage() FrontPage {
	var frontPage FrontPage
	frontPage.Page = Page{}
	return frontPage
}

func main() {
	handleFlags()

	var err error
	conf, err = config.New()
	if err != nil {
		log.Fatal("Could not get configuration: ", err)
	}
	imagePath := path.Join(conf.GetString("configPath"), "images")
	fontPath := path.Join(conf.GetString("webRoot"), "static", "fonts")

	router := httprouter.New()

	router.GET("/", frontPageHandler)

	router.ServeFiles("/images/*filepath", http.Dir(imagePath))
	router.ServeFiles("/static/fonts/*filepath", http.Dir(fontPath))

	router.GET("/static/css/*filePath", cssHandler)
	router.GET("/static/js/*filePath", jsHandler)

	log.Fatal(http.ListenAndServe(":8080", router))
}

func handleFlags(){
	ver := flag.Bool("version", false, "Prints the version")
	flag.Parse()

	if *ver {
		fmt.Println(version)
		os.Exit(0)
	}
}

func jsHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Add("Content-Type", "application/javascript")
	path := ps.ByName("filePath")
	root := conf.GetString("webRoot")
	http.ServeFile(w, r, root + "/static/js"+path)
}

func cssHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Add("Content-Type", "text/css")
	path := ps.ByName("filePath")
	root := conf.GetString("webRoot")
	http.ServeFile(w, r, root + "/static/css/"+path)
}

func frontPageHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	root := conf.GetString("webRoot")
	http.ServeFile(w, r, root + "/front.html")
}

var myClient = &http.Client{Timeout: 10 * time.Second}

func checkError(e error, w http.ResponseWriter) {
	if e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
	}
}

func getJSON(url string, target interface{}) error {
	r, err := myClient.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, &target)
	if err != nil {
		return err
	}

	return nil
}
