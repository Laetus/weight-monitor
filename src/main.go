package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
)

type Entry struct {
	Weight float64   `json:"weight"`
	Date   time.Time `json:"date,omitempty"`
}

type Config struct {
	OauthClientId string `split_words:"true"`
}

type IndexData struct {
	Config  Config
	Entries []Entry
}

func trimOauthClientId() {
	splits := strings.Split(config.OauthClientId, "/")
	if len(splits) == 1 {
		return
	}
	config.OauthClientId = splits[len(splits)-1]
}

var entries []Entry
var config Config

func saveEntry(entry Entry) {
	// Limit entries length to 10 to avoid pagination and memory issues
	if len(entries) > 10 {
		_, entries = entries[0], entries[1:]
	}
	entries = append(entries, entry)

}

func homeLink(w http.ResponseWriter, r *http.Request) {
	// path is relative to project root
	const templatePath string = "templates/index.html"
	tmpl, err := template.ParseFiles(templatePath)

	if err != nil || tmpl == nil {
		fmt.Fprintf(w, "Something went wrong :-(")
		return
	}

	keys, ok := r.URL.Query()["weight"]
	if ok && len(keys) == 1 {
		weight, err := strconv.ParseFloat(keys[0], 64)
		if err == nil {
			log.Println(weight)
			entry := Entry{weight, time.Now()}
			saveEntry(entry)
		}
	}

	sort.Slice(entries, func(a, b int) bool {
		return entries[a].Date.After(entries[b].Date)
	})
	w.Header().Add("Content-Type", "text/html")
	tmpl.Execute(w, IndexData{config, entries})
}

func validate(r *http.Request) bool {
	token := r.Header.Get("Authorization")
	if token == "" {
		return false
	}

	return strings.Replace(token, "Bearer ", "", 1) != ""
}

func createWeight(w http.ResponseWriter, r *http.Request) {
	if !validate(r) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(r.Body)

	entry := Entry{-1, time.Now()}

	err := decoder.Decode(&entry)

	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
	} else {
		saveEntry(entry)
	}

	json.NewEncoder(w).Encode(entries)
}

func listWeights(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(entries)
}

func commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func main() {
	entries = []Entry{}

	err := envconfig.Process("", &config)
	if err != nil {
		log.Println("Config could not be loaded.")
		log.Fatal(err.Error())
	}

	router := mux.NewRouter().StrictSlash(true)
	router.Use(commonMiddleware)
	router.HandleFunc("/", homeLink)
	router.HandleFunc("/weight", listWeights).Methods("GET")
	router.HandleFunc("/weight", createWeight).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", router))
}
