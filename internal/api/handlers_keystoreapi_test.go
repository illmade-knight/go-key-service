// REFACTOR: This test file is updated to validate the fully refactored,
// URN-aware API handlers.

package api_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/illmade-knight/go-key-service/internal/api"
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
		rr := httptest.NewRecorder()

		// Act
		apiHandler.StoreKeyHandler(rr, req)

		// Assert
		assert.Equal(t, http.StatusCreated, rr.Code)
		mockStore.AssertExpectations(t)
	})

	t.Run("Failure - Invalid URN", func(t *testing.T) {
		// Arrange
		mockStore := new(MockStore) // No expectations, as it shouldn't be called.
		apiHandler := &api.API{Store: mockStore, Logger: logger}
		req := httptest.NewRequest(http.MethodPost, "/keys/invalid-urn", bytes.NewReader([]byte(testKey)))
		req.SetPathValue("entityURN", "invalid-urn")
		rr := httptest.NewRecorder()

		// Act
		apiHandler.StoreKeyHandler(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
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
		assert.Equal(t, http.StatusNotFound, rr.Code)
		mockStore.AssertExpectations(t)
	})
}
