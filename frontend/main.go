package main

import (
	"encoding/json"
	"github.com/bhutch29/abv/model"
	"github.com/julienschmidt/httprouter"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// Page is the backing type for all pages
type Page struct {
	MainWindowHTML template.HTML
}

// FrontPage is the primary view object
type FrontPage struct {
	Page
	Drinks []model.StockedDrink
}

func newPage() Page {
	var page Page
	page.MainWindowHTML = template.HTML(stringFromFile("frontPage.html"))
	return page
}

func stringFromFile(path string) string {
	b, _ := ioutil.ReadFile(path)
	return string(b)
}

func newFrontPage() FrontPage {
	var frontPage FrontPage
	frontPage.Page = newPage()
	return frontPage
}

var templates = template.Must(template.ParseFiles("front.html"))

func main() {
	router := httprouter.New()

	router.GET("/", frontPageHandler)
	router.GET("/static/css/*filePath", cssHandler)
	router.GET("/static/js/*filePath", jsHandler)

	log.Fatal(http.ListenAndServe(":8080", router))
}

func jsHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Add("Content-Type", "application/javascript")
	path := ps.ByName("filePath")
	http.ServeFile(w, r, "static/js"+path)
}

func cssHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Add("Content-Type", "text/css")
	path := ps.ByName("filePath")
	http.ServeFile(w, r, "static/css/"+path)
}

func frontPageHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	p := newFrontPage()
	err := getJSON("http://localhost:8081/inventory", &p.Drinks)
	checkError(err, w)
	err = templates.ExecuteTemplate(w, "front.html", p)
	checkError(err, w)
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
