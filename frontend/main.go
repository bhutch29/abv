package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/bhutch29/abv/config"
	"github.com/bhutch29/abv/model"
	"github.com/julienschmidt/httprouter"
	"github.com/spf13/viper"
)

var (
	conf    *viper.Viper
	version = "undefined"
	tmpl    *template.Template
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
	webRoot := conf.GetString("webRoot")
	fontPath := path.Join(webRoot, "static", "fonts")

	frontHTML := path.Join(webRoot, "front.html")
	tmpl = template.Must(template.ParseFiles(frontHTML))

	router := httprouter.New()

	router.GET("/", frontPageHandler)

	router.ServeFiles("/images/*filepath", http.Dir(imagePath))
	router.ServeFiles("/static/fonts/*filepath", http.Dir(fontPath))

	router.GET("/static/css/*filePath", cssHandler)
	router.GET("/static/js/*filePath", jsHandler)
	router.GET("/static/html/*filePath", htmlHandler)

	log.Fatal(http.ListenAndServe(":8080", router))
}

func handleFlags() {
	ver := flag.Bool("version", false, "Prints the version")
	flag.Parse()

	if *ver {
		fmt.Println(version)
		os.Exit(0)
	}
}

func jsHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Add("Content-Type", "application/javascript")
	filePath := ps.ByName("filePath")
	root := conf.GetString("webRoot")
	fullPath := path.Join(root, "static", "js", filePath)
	http.ServeFile(w, r, fullPath)
}

func cssHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Add("Content-Type", "text/css")
	filePath := ps.ByName("filePath")
	root := conf.GetString("webRoot")
	fullPath := path.Join(root, "static", "css", filePath)
	http.ServeFile(w, r, fullPath)
}

func htmlHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Add("Content-Type", "text/html")
	filePath := ps.ByName("filePath")
	root := conf.GetString("webRoot")
	fullPath := path.Join(root, "static", "html", filePath)
	http.ServeFile(w, r, fullPath)
}

func frontPageHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	apiURL := conf.GetString("apiUrl")
	tmpl.Execute(w, apiURL)
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
