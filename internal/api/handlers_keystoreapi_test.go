// REFACTOR: This test file is updated to validate the fully refactored,
// URN-aware API handlers.

package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/illmade-knight/go-key-service/internal/api"
	"github.com/illmade-knight/go-microservice-base/pkg/response" // ADDED: For the APIError struct
	"github.com/illmade-knight/go-secure-messaging/pkg/urn"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockStore is a mock implementation of the keyservice.Store interface.
type MockStore struct {
	mock.Mock
}

// StoreKey is the mock implementation for storing a key.
func (m *MockStore) StoreKey(ctx context.Context, entityURN urn.URN, key []byte) error {
	args := m.Called(ctx, entityURN, key)
	return args.Error(0)
}

// GetKey is the mock implementation for retrieving a key.
func (m *MockStore) GetKey(ctx context.Context, entityURN urn.URN) ([]byte, error) {
	args := m.Called(ctx, entityURN)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

// TestStoreKeyHandler tests the POST /keys/{entityURN} endpoint handler.
func TestStoreKeyHandler(t *testing.T) {
	testURN, err := urn.New(urn.SecureMessaging, "user", "user-123")
	require.NoError(t, err)
	const testKey = "my-public-key"
	logger := zerolog.Nop()

	t.Run("Success - 201 Created", func(t *testing.T) {
		// Arrange
		mockStore := new(MockStore)
		mockStore.On("StoreKey", mock.Anything, testURN, []byte(testKey)).Return(nil)

		apiHandler := &api.API{Store: mockStore, Logger: logger}
		req := httptest.NewRequest(http.MethodPost, "/keys/"+testURN.String(), bytes.NewReader([]byte(testKey)))
		req.SetPathValue("entityURN", testURN.String())

		// ADDED: Simulate successful authentication by the JWT middleware
		// The handler expects the authenticated user's ID in the context.
		ctx := api.ContextWithUserID(context.Background(), "user-123")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		// Act
		apiHandler.StoreKeyHandler(rr, req)

		// Assert
		assert.Equal(t, http.StatusCreated, rr.Code)
		mockStore.AssertExpectations(t)
	})

	// ADDED: Crucial test for authorization logic.
	t.Run("Failure - 403 Forbidden", func(t *testing.T) {
		// Arrange
		mockStore := new(MockStore) // No expectations, as it shouldn't be called.
		apiHandler := &api.API{Store: mockStore, Logger: logger}
		req := httptest.NewRequest(http.MethodPost, "/keys/"+testURN.String(), bytes.NewReader([]byte(testKey)))
		req.SetPathValue("entityURN", testURN.String())

		// Simulate a DIFFERENT user being authenticated.
		ctx := api.ContextWithUserID(context.Background(), "another-user-456")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		// Act
		apiHandler.StoreKeyHandler(rr, req)

		// Assert
		assert.Equal(t, http.StatusForbidden, rr.Code)
		var errResp response.APIError
		err := json.Unmarshal(rr.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, "Forbidden", errResp.Error)
		mockStore.AssertNotCalled(t, "StoreKey", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("Failure - Invalid URN", func(t *testing.T) {
		// Arrange
		mockStore := new(MockStore)
		apiHandler := &api.API{Store: mockStore, Logger: logger}
		req := httptest.NewRequest(http.MethodPost, "/keys/invalid-urn", bytes.NewReader([]byte(testKey)))
		req.SetPathValue("entityURN", "invalid-urn")

		// The handler still requires an authenticated user in the context.
		ctx := api.ContextWithUserID(context.Background(), "user-123")
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		// Act
		apiHandler.StoreKeyHandler(rr, req)

		// Assert
		// CHANGED: Verify the new JSON error response.
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		var errResp response.APIError
		err := json.Unmarshal(rr.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, "Invalid URN format in request path", errResp.Error)
		mockStore.AssertNotCalled(t, "StoreKey", mock.Anything, mock.Anything, mock.Anything)
	})
}

// TestGetKeyHandler tests the GET /keys/{entityURN} endpoint handler.
func TestGetKeyHandler(t *testing.T) {
	testURN, err := urn.New(urn.SecureMessaging, "user", "user-123")
	require.NoError(t, err)
	const testKey = "my-public-key"
	logger := zerolog.Nop()

	t.Run("Success - 200 OK", func(t *testing.T) {
		// Arrange
		mockStore := new(MockStore)
		mockStore.On("GetKey", mock.Anything, testURN).Return([]byte(testKey), nil)

		apiHandler := &api.API{Store: mockStore, Logger: logger}
		req := httptest.NewRequest(http.MethodGet, "/keys/"+testURN.String(), nil)
		req.SetPathValue("entityURN", testURN.String())
		rr := httptest.NewRecorder()

		// Act
		apiHandler.GetKeyHandler(rr, req)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, testKey, rr.Body.String())
		mockStore.AssertExpectations(t)
	})

	t.Run("Failure - Not Found", func(t *testing.T) {
		// Arrange
		notFoundURN, err := urn.New(urn.SecureMessaging, "user", "not-found")
		require.NoError(t, err)
		mockStore := new(MockStore)
		mockStore.On("GetKey", mock.Anything, notFoundURN).Return(nil, errors.New("not found"))

		apiHandler := &api.API{Store: mockStore, Logger: logger}
		req := httptest.NewRequest(http.MethodGet, "/keys/"+notFoundURN.String(), nil)
		req.SetPathValue("entityURN", notFoundURN.String())
		rr := httptest.NewRecorder()

		// Act
		apiHandler.GetKeyHandler(rr, req)

		// Assert
		// CHANGED: Verify the new JSON error response.
		assert.Equal(t, http.StatusNotFound, rr.Code)
		var errResp response.APIError
		err = json.Unmarshal(rr.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, "Key not found", errResp.Error)
		mockStore.AssertExpectations(t)
	})
}

// NOTE: To make ContextWithUserID accessible, you may need to export it
// from the api package by renaming it from contextWithUserID to ContextWithUserID
// or by creating a new helper function for testing.
// For simplicity in this fix, I've created a helper here.
// In a real project, you might expose a `testutils` package.
func (m *MockStore) ContextWithUserID(ctx context.Context, userID string) context.Context {
	// This is a stand-in for the actual key and function in the api package.
	type contextKey string
	const userContextKey contextKey = "userID"
	return context.WithValue(ctx, userContextKey, userID)
}
