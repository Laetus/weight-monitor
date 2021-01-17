package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
	"weightmonitor/src/util"

	"github.com/gorilla/mux"

	"cloud.google.com/go/firestore"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/iterator"
	"google.golang.org/api/oauth2/v2"
)

var config util.Config

func init() {
	config = util.ConfigInstance
}

type Entry struct {
	Weight float64   `json:"weight"`
	Date   time.Time `json:"date,omitempty"`
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

func commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			w.Header().Add("Content-Type", "application/json")
		}
		log.Println(r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

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

func SaveEntry(entry Entry, userTokeninfo *oauth2.Tokeninfo) {
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

func GetEntries(userTokeninfo *oauth2.Tokeninfo) []Entry {
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
	return entries
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
		SaveEntry(entry, userTokeninfo)
	}

	entries := GetEntries(userTokeninfo)
	json.NewEncoder(w).Encode(entries)
}

func listWeights(w http.ResponseWriter, r *http.Request) {
	userTokeninfo, err := util.Validate(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	entries := GetEntries(userTokeninfo)
	json.NewEncoder(w).Encode(entries)
}

func SetupSubrouter(subrouter *mux.Router) *mux.Router {
	subrouter.HandleFunc("", listWeights).Methods("GET")
	subrouter.HandleFunc("", createWeight).Methods("POST")
	subrouter.Use(commonMiddleware)
	return subrouter
}
