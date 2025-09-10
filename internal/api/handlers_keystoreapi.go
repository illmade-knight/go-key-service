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

// StoreKeyHandler manages the POST requests for user keys.
func (a *API) StoreKeyHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userID")
	a.Logger.Log().Str("user_id", userID).Str("method", "POST").Msg("updating key")
	key, err := io.ReadAll(r.Body)
	if err != nil {
		a.Logger.Warn().Err(err).Msg("Cannot read request body")
		http.Error(w, "Cannot read request body", http.StatusBadRequest)
		return
	}
	if err := a.Store.StoreKey(userID, key); err != nil {
		a.Logger.Error().Err(err).Msg("Failed to store key")
		http.Error(w, "Failed to store key", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	a.Logger.Info().Msg("Stored public key")
}

// GetKeyHandler manages GET  requests for user keys.
func (a *API) GetKeyHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userID")
	logger := a.Logger.With().Str("user_id", userID).Str("method", r.Method).Logger()

	key, err := a.Store.GetKey(userID)
	if err != nil {
		logger.Warn().Msg("Key not found")
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	_, _ = w.Write(key)
	logger.Info().Msg("Served public key")

}
