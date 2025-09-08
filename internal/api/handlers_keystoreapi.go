// Package api contains the private HTTP handlers for the key service.
package api

import (
	"io"
	"net/http"

	"github.com/illmade-knight/go-key-service/pkg/keyservice"
	"github.com/rs/zerolog"
)

// API holds the dependencies for the HTTP handlers.
type API struct {
	Store  keyservice.Store
	Logger zerolog.Logger
}

// KeyHandler manages GET and POST requests for user keys.
func (a *API) KeyHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Path[len("/keys/"):]
	logger := a.Logger.With().Str("user_id", userID).Str("method", r.Method).Logger()

	switch r.Method {
	case http.MethodPost:
		key, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Warn().Err(err).Msg("Cannot read request body")
			http.Error(w, "Cannot read request body", http.StatusBadRequest)
			return
		}
		if err := a.Store.StoreKey(userID, key); err != nil {
			logger.Error().Err(err).Msg("Failed to store key")
			http.Error(w, "Failed to store key", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		logger.Info().Msg("Stored public key")

	case http.MethodGet:
		key, err := a.Store.GetKey(userID)
		if err != nil {
			logger.Warn().Msg("Key not found")
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write(key)
		logger.Info().Msg("Served public key")

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
