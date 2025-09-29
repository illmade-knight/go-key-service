package api

import (
	"context"
	"io"
	"net/http"

	"github.com/illmade-knight/go-key-service/pkg/keyservice"
	"github.com/illmade-knight/go-microservice-base/pkg/response" // ADDED: Import the new response helper
	"github.com/illmade-knight/go-secure-messaging/pkg/urn"
	"github.com/rs/zerolog"
)

// API now includes the JWTSecret for the middleware.
type API struct {
	Store     keyservice.Store
	Logger    zerolog.Logger
	JWTSecret string
}

type contextKey string

// UserContextKey is the key used to store the authenticated user's ID from the JWT.
const UserContextKey contextKey = "userID"

// GetUserIDFromContext safely retrieves the user ID from the request context.
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserContextKey).(string)
	return userID, ok
}

// ContextWithUserID is a helper function for tests to inject a user ID
// into a context, simulating a successful authentication from middleware.
func ContextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserContextKey, userID)
}

// StoreKeyHandler manages the POST requests for entity keys.
func (a *API) StoreKeyHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Get the authenticated user's ID securely from the JWT context.
	authedUserID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		a.Logger.Error().Msg("User ID not found in context; middleware may be misconfigured.")
		// CHANGED: Use standardized JSON error response
		response.WriteJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// 2. Get the target URN from the URL path.
	entityURNStr := r.PathValue("entityURN")
	entityURN, err := urn.Parse(entityURNStr)
	if err != nil {
		a.Logger.Warn().Err(err).Str("raw_urn", entityURNStr).Msg("Invalid URN format in request path")
		// CHANGED: Use standardized JSON error response
		response.WriteJSONError(w, http.StatusBadRequest, "Invalid URN format in request path")
		return
	}

	// 3. THE CRITICAL SECURITY CHECK:
	if authedUserID != entityURN.EntityID() {
		a.Logger.Warn().Str("authed_user", authedUserID).Str("target_urn", entityURN.String()).Msg("Authorization failed: User attempted to store key for another entity.")
		// CHANGED: Use standardized JSON error response
		response.WriteJSONError(w, http.StatusForbidden, "Forbidden")
		return
	}

	logger := a.Logger.With().Str("entity_urn", entityURN.String()).Logger()
	key, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to read request body")
		response.WriteJSONError(w, http.StatusBadRequest, "Cannot read request body")
		return
	}

	logger.Info().Int("byteLength", len(key)).Msg("[Checkpoint 2: RECEIPT] Key received from client")

	if err := a.Store.StoreKey(r.Context(), entityURN, key); err != nil {
		logger.Error().Err(err).Msg("Failed to store key")
		// CHANGED: Use standardized JSON error response
		response.WriteJSONError(w, http.StatusInternalServerError, "Failed to store key")
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
		// CHANGED: Use standardized JSON error response
		response.WriteJSONError(w, http.StatusBadRequest, "Invalid URN format")
		return
	}

	logger := a.Logger.With().Str("entity_urn", entityURN.String()).Logger()
	key, err := a.Store.GetKey(r.Context(), entityURN)
	if err != nil {
		logger.Warn().Err(err).Msg("Key not found")
		// CHANGED: Use standardized JSON error response
		response.WriteJSONError(w, http.StatusNotFound, "Key not found")
		return
	}

	logger.Info().Int("byteLength", len(key)).Msg("[Checkpoint 3: RETRIEVAL] Key retrieved from store to be sent")

	w.Header().Set("Content-Type", "application/octet-stream")
	_, err = w.Write(key)
	if err != nil {
		logger.Error().Err(err).Msg("write fail")
	}
	logger.Info().Msg("Successfully retrieved public key")
}
