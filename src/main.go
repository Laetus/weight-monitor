package main

import (
	"fmt"
	"log"
	"net/http"
	"text/template"
	"weightmonitor/src/api"
	"weightmonitor/src/util"

	"github.com/gorilla/mux"
)

type IndexData struct {
	Config  util.Config
	Entries []api.Entry
}

var config util.Config

func homeLink(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	// path is relative to project root
	const templatePath string = "templates/index.html"
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil || tmpl == nil {
		fmt.Fprintf(w, "Something went wrong :-(")
		return
	}

	userTokeninfo, authErr := util.Validate(r)
	var entries []api.Entry = make([]api.Entry, 0)
	if authErr == nil {
		log.Println("load entries for user")
		entries = api.GetEntries(userTokeninfo)
	}
	tmpl.Execute(w, IndexData{config, entries})
}

func init() {
	config = util.ConfigInstance
}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	api.SetupSubrouter(router.PathPrefix("/weight").Subrouter())
	router.HandleFunc("/", homeLink)
	log.Fatal(http.ListenAndServe(":8080", router))
}
