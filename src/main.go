package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type Entry struct {
	Weight float64   `json:"weight"`
	Date   time.Time `json:"date,omitempty"`
}

var entries []Entry

func homeLink(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome home!")
}

func createWeight(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	entry := Entry{-1, time.Now()}

	err := decoder.Decode(&entry)

	if err != nil {
		log.Println(err)
	}
	log.Println(entry)

	// Limit entries length to 10 to avoid pagination and memory issues
	if len(entries) > 10 {
		_, entries = entries[0], entries[1:]
	}
	entries = append(entries, entry)
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
	router := mux.NewRouter().StrictSlash(true)
	router.Use(commonMiddleware)
	router.HandleFunc("/", homeLink)
	router.HandleFunc("/weight", listWeights).Methods("GET")
	router.HandleFunc("/weight", createWeight).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", router))
}
