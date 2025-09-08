package main

import (
	"io"
	"log"
	"net/http"

	"github.com/illmade-knight/go-key-service/pkg/storage"
)

// api holds the dependencies for our service, like the storage layer.
type api struct {
	store storage.Store
}

func (a *api) keyHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Path[len("/keys/"):]

	switch r.Method {
	case http.MethodPost:
		key, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Cannot read request body", http.StatusBadRequest)
			return
		}
		if err := a.store.StoreKey(userID, key); err != nil {
			http.Error(w, "Failed to store key", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		log.Printf("Stored public key for user: %s", userID)

	case http.MethodGet:
		key, err := a.store.GetKey(userID)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(key)
		log.Printf("Served public key for user: %s", userID)
	}
}

func main() {
	// Create the storage implementation
	store := storage.NewInMemoryStore()

	// Inject the dependency into our api struct
	app := &api{store: store}

	http.HandleFunc("/keys/", app.keyHandler)
	log.Println("Key Service listening on :8081...")
	http.ListenAndServe(":8081", nil)
}
