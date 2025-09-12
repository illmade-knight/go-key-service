// REFACTOR: This file is updated to handle URNs in the URL path. It now
// parses and validates the incoming entity identifier before passing it to the
// storage layer and uses the request's context for all downstream calls.

// Package api contains the private HTTP handlers for the key service.
package api

import (
	"io"
	"net/http"

	"github.com/illmade-knight/go-key-service/pkg/keyservice"
	"github.com/illmade-knight/go-secure-messaging/pkg/urn"
	"github.com/rs/zerolog"
)

// API holds the dependencies for the HTTP handlers.
type API struct {
	Store  keyservice.Store
	Logger zerolog.Logger
}

// StoreKeyHandler manages the POST requests for entity keys.
func (a *API) StoreKeyHandler(w http.ResponseWriter, r *http.Request) {
	entityURNStr := r.PathValue("entityURN")
	entityURN, err := urn.Parse(entityURNStr)
	if err != nil {
		a.Logger.Warn().Err(err).Str("raw_urn", entityURNStr).Msg("Invalid URN format in request path")
		http.Error(w, "Invalid URN format in request path", http.StatusBadRequest)
		return
	}

	logger := a.Logger.With().Str("entity_urn", entityURN.String()).Str("method", r.Method).Logger()
	logger.Debug().Msg("Storing key for entity")

	key, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Warn().Err(err).Msg("Cannot read request body")
		http.Error(w, "Cannot read request body", http.StatusBadRequest)
		return
	}
	if err := a.Store.StoreKey(r.Context(), entityURN, key); err != nil {
		logger.Error().Err(err).Msg("Failed to store key")
		http.Error(w, "Failed to store key", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	logger.Info().Msg("Successfully stored public key")
}

// GetKeyHandler manages GET requests for entity keys.
func (a *API) GetKeyHandler(w http.ResponseWriter, r *http.Request) {
	entityURNStr := r.PathValue("entityURN")
	entityURN, err := urn.Parse(entityURNStr)
	if err != nil {
		a.Logger.Warn().Err(err).Str("raw_urn", entityURNStr).Msg("Invalid URN format in request path")
		http.Error(w, "Invalid URN format in request path", http.StatusBadRequest)
		return
	}

	logger := a.Logger.With().Str("entity_urn", entityURN.String()).Str("method", r.Method).Logger()
	logger.Debug().Msg("Retrieving key for entity")

	key, err := a.Store.GetKey(r.Context(), entityURN)
	if err != nil {
		logger.Warn().Err(err).Msg("Key not found")
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	_, _ = w.Write(key)
	logger.Info().Msg("Successfully served public key")
}
