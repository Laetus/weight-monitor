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
	"google.golang.org/api/oauth2/v2"
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

func validate(r *http.Request) (*oauth2.Tokeninfo, error) {
	idToken := r.Header.Get("Authorization")

	idToken = strings.Replace(idToken, "Bearer ", "", 1)

	oauth2Service, err := oauth2.New(httpClient)
	tokenInfoCall := oauth2Service.Tokeninfo()
	tokenInfoCall.IdToken(idToken)
	tokenInfo, err := tokenInfoCall.Do()
	return tokenInfo, err
}

var entries []Entry
var config Config
var httpClient = &http.Client{}

func saveEntry(entry Entry) {
	sortEntries()
	// Limit entries length to 10 to avoid pagination and memory issues
	if len(entries) > 9 {
		entries = entries[:len(entries)-1]
	}
	entries = append(entries, entry)

}

func sortEntries() {
	sort.Slice(entries, func(a, b int) bool {
		return entries[a].Date.After(entries[b].Date)
	})
}

func homeLink(w http.ResponseWriter, r *http.Request) {
	// path is relative to project root
	const templatePath string = "templates/index.html"
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil || tmpl == nil {
		fmt.Fprintf(w, "Something went wrong :-(")
		return
	}

	_, authErr := validate(r)
	if authErr != nil {
		w.Header().Add("Content-Type", "text/html")
		tmpl.Execute(w, IndexData{config, make([]Entry, 0)})
		return
	}

	keys, ok := r.URL.Query()["weight"]
	if ok && len(keys) == 1 {
		weight, err := strconv.ParseFloat(keys[0], 64)
		if err == nil {
			entry := Entry{weight, time.Now()}
			saveEntry(entry)
		}
	}
	sortEntries()
	w.Header().Add("Content-Type", "text/html")
	tmpl.Execute(w, IndexData{config, entries})
}

func createWeight(w http.ResponseWriter, r *http.Request) {
	_, err := validate(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(r.Body)

	entry := Entry{-1, time.Now()}

	err = decoder.Decode(&entry)

	log.Println(entry.Weight)
	if err != nil || entry.Weight <= 0 {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
	} else {
		saveEntry(entry)
	}

	json.NewEncoder(w).Encode(entries)
}

func listWeights(w http.ResponseWriter, r *http.Request) {
	_, err := validate(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	sortEntries()
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
