package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"text/template"
	"time"
	"weightmonitor/src/util"

	"cloud.google.com/go/firestore"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/iterator"
	"google.golang.org/api/oauth2/v2"
)

type Entry struct {
	Weight float64   `json:"weight"`
	Date   time.Time `json:"date,omitempty"`
}

type IndexData struct {
	Config  util.Config
	Entries []Entry
}

func createClient(ctx context.Context) *firestore.Client {
	credentials, error := google.FindDefaultCredentials(ctx, compute.ComputeScope)
	if error != nil {
		log.Println(error)
	}
	projectID := credentials.ProjectID

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Printf("Failed to create client: %v", err)
	}
	return client
}

var config util.Config
var httpClient = &http.Client{}
var firestoreClient = createClient(context.Background())

func getUserDoc(userTokeninfo *oauth2.Tokeninfo) *firestore.DocumentRef {

	id := userTokeninfo.UserId
	if id == "" {
		id = "anonymous"
	}

	ctx := context.Background()
	_, err := firestoreClient.Collection("users").Doc(id).Set(ctx, map[string]interface{}{
		"Audience":      userTokeninfo.Audience,
		"Email":         userTokeninfo.Email,
		"UserId":        userTokeninfo.UserId,
		"Scope":         userTokeninfo.Scope,
		"VerifiedEmail": userTokeninfo.VerifiedEmail,
	})

	if err != nil {
		log.Printf("failed to get user reference for user %s: \n %s", id, err)
		return nil
	}
	return firestoreClient.Collection("users").Doc(id)

}

func saveEntry(entry Entry, userTokeninfo *oauth2.Tokeninfo) {
	// Limit entries length to 10 to avoid pagination and memory issues
	ctx := context.Background()
	ref := getUserDoc(userTokeninfo)
	if ref == nil {
		log.Println("failed getting user")
		return
	}
	_, err := ref.Collection("entries").Doc(entry.Date.Format("2006-01-02 15:04:05")).Set(ctx, entry)

	if err != nil {
		log.Println("failed to save entry")
	}
	log.Println("saved entry")

}

func getEntries(userTokeninfo *oauth2.Tokeninfo) []Entry {
	entries := []Entry{}
	ref := getUserDoc(userTokeninfo)

	if ref == nil {
		log.Println("failed getting user")
		return nil
	}

	query := ref.Collection("entries").OrderBy("Date", firestore.Desc).Limit(config.EntriesLimit)
	ctx := context.Background()
	iter := query.Documents(ctx)
	for {
		doc, err := iter.Next()
		if err != nil {
			if err != iterator.Done {
				log.Println(err)
			}
			break
		}
		var entry Entry
		doc.DataTo(&entry)
		entries = append(entries, entry)
	}
	log.Printf("finished loading entries: %s", entries)
	return entries
}

func homeLink(w http.ResponseWriter, r *http.Request) {
	// path is relative to project root
	const templatePath string = "templates/index.html"
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil || tmpl == nil {
		fmt.Fprintf(w, "Something went wrong :-(")
		return
	}

	userTokeninfo, authErr := util.Validate(r)
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
			saveEntry(entry, userTokeninfo)
		}
	}
	entries := getEntries(userTokeninfo)
	w.Header().Add("Content-Type", "text/html")
	tmpl.Execute(w, IndexData{config, entries})
}

func createWeight(w http.ResponseWriter, r *http.Request) {
	userTokeninfo, err := util.Validate(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(r.Body)

	entry := Entry{-1, time.Now()}

	err = decoder.Decode(&entry)

	if err != nil || entry.Weight <= 0 {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
	} else {
		saveEntry(entry, userTokeninfo)
	}

	entries := getEntries(userTokeninfo)
	json.NewEncoder(w).Encode(entries)
}

func listWeights(w http.ResponseWriter, r *http.Request) {
	userTokeninfo, err := util.Validate(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	entries := getEntries(userTokeninfo)
	json.NewEncoder(w).Encode(entries)
}

func commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			w.Header().Add("Content-Type", "application/json")
		}
		log.Println(r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func main() {
	config = util.ConfigInstance
	router := mux.NewRouter().StrictSlash(true)
	router.Use(commonMiddleware)
	router.HandleFunc("/", homeLink)
	router.HandleFunc("/weight", listWeights).Methods("GET")
	router.HandleFunc("/weight", createWeight).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", router))
}
