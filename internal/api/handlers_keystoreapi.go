package api

import (
	"io"
	"net/http"

	"github.com/illmade-knight/go-key-service/pkg/keyservice"
	"github.com/illmade-knight/go-secure-messaging/pkg/urn"
	"github.com/rs/zerolog"
)

// API now includes the JWTSecret for the middleware.
type API struct {
	Store     keyservice.Store
	Logger    zerolog.Logger
	JWTSecret string
}

// StoreKeyHandler manages the POST requests for entity keys.
func (a *API) StoreKeyHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Get the authenticated user's ID securely from the JWT context.
	authedUserID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		a.Logger.Error().Msg("User ID not found in context; middleware may be misconfigured.")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 2. Get the target URN from the URL path.
	entityURNStr := r.PathValue("entityURN")
	entityURN, err := urn.Parse(entityURNStr)
	if err != nil {
		a.Logger.Warn().Err(err).Str("raw_urn", entityURNStr).Msg("Invalid URN format in request path")
		http.Error(w, "Invalid URN format in request path", http.StatusBadRequest)
		return
	}

	// 3. THE CRITICAL SECURITY CHECK:
	// Ensure the authenticated user is only trying to store a key for themselves.
	// The ID from the token (`sub` claim) must match the ID in the URN.
	if authedUserID != entityURN.EntityID() {
		a.Logger.Warn().Str("authed_user", authedUserID).Str("target_urn", entityURN.String()).Msg("Authorization failed: User attempted to store key for another entity.")
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	logger := a.Logger.With().Str("entity_urn", entityURN.String()).Logger()
	key, err := io.ReadAll(r.Body)
	if err != nil {
		// ...
		return
	}

	// --- ADD THIS LOGGING ---
	logger.Info().Int("byteLength", len(key)).Msg("[Checkpoint 2: RECEIPT] Key received from client")
	// -------------------------

	if err := a.Store.StoreKey(r.Context(), entityURN, key); err != nil {
		logger.Error().Err(err).Msg("Failed to store key")
		http.Error(w, "Failed to store key", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	logger.Info().Msg("Successfully stored public key")
}

// GetKeyHandler remains public as clients need to fetch others' public keys.
func (a *API) GetKeyHandler(w http.ResponseWriter, r *http.Request) {
	entityURNStr := r.PathValue("entityURN")
	entityURN, err := urn.Parse(entityURNStr)
	if err != nil {
		a.Logger.Warn().Err(err).Str("raw_urn", entityURNStr).Msg("Invalid URN format")
		http.Error(w, "Invalid URN format", http.StatusBadRequest)
		return
	}

	logger := a.Logger.With().Str("entity_urn", entityURN.String()).Logger()
	key, err := a.Store.GetKey(r.Context(), entityURN)
	if err != nil {
		logger.Warn().Err(err).Msg("Key not found")
		http.NotFound(w, r)
		return
	}
	
	logger.Info().Int("byteLength", len(key)).Msg("[Checkpoint 3: RETRIEVAL] Key retrieved from store to be sent")

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(key)
	logger.Info().Msg("Successfully retrieved public key")
}
