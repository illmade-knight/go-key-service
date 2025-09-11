package api_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/illmade-knight/go-key-service/internal/api"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStore is a mock implementation of the keyservice.Store interface
// generated using testify/mock.
type MockStore struct {
	mock.Mock
}

// StoreKey is the mock implementation for storing a key.
func (m *MockStore) StoreKey(userID string, key []byte) error {
	args := m.Called(userID, key)
	return args.Error(0)
}

// GetKey is the mock implementation for retrieving a key.
func (m *MockStore) GetKey(userID string) ([]byte, error) {
	args := m.Called(userID)
	// Handle nil case for the byte slice if the first argument isn't one.
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

// TestStoreKeyHandler tests the POST /keys/{userID} endpoint handler.
func TestStoreKeyHandler(t *testing.T) {
	const testUserID = "user-123"
	const testKey = "my-public-key"
	logger := zerolog.Nop()

	t.Run("Success - 201 Created", func(t *testing.T) {
		// Arrange
		mockStore := new(MockStore)
		mockStore.On("StoreKey", testUserID, []byte(testKey)).Return(nil)

		apiHandler := &api.API{Store: mockStore, Logger: logger}
		req := httptest.NewRequest(http.MethodPost, "/keys/"+testUserID, bytes.NewReader([]byte(testKey)))
		req.SetPathValue("userID", testUserID) // Set the path value for the test request
		rr := httptest.NewRecorder()

		// Act
		apiHandler.StoreKeyHandler(rr, req)

		// Assert
		assert.Equal(t, http.StatusCreated, rr.Code)
		mockStore.AssertExpectations(t)
	})

	t.Run("Store Fails - 500 Internal Server Error", func(t *testing.T) {
		// Arrange
		mockStore := new(MockStore)
		mockStore.On("StoreKey", testUserID, []byte(testKey)).Return(errors.New("database is down"))

		apiHandler := &api.API{Store: mockStore, Logger: logger}
		req := httptest.NewRequest(http.MethodPost, "/keys/"+testUserID, bytes.NewReader([]byte(testKey)))
		req.SetPathValue("userID", testUserID) // Set the path value for the test request
		rr := httptest.NewRecorder()

		// Act
		apiHandler.StoreKeyHandler(rr, req)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		mockStore.AssertExpectations(t)
	})
}

// TestGetKeyHandler tests the GET /keys/{userID} endpoint handler.
func TestGetKeyHandler(t *testing.T) {
	const testUserID = "user-123"
	const testKey = "my-public-key"
	logger := zerolog.Nop()

	t.Run("Success - 200 OK", func(t *testing.T) {
		// Arrange
		mockStore := new(MockStore)
		mockStore.On("GetKey", testUserID).Return([]byte(testKey), nil)

		apiHandler := &api.API{Store: mockStore, Logger: logger}
		req := httptest.NewRequest(http.MethodGet, "/keys/"+testUserID, nil)
		req.SetPathValue("userID", testUserID) // Set the path value for the test request
		rr := httptest.NewRecorder()

		// Act
		apiHandler.GetKeyHandler(rr, req)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, testKey, rr.Body.String())
		assert.Equal(t, "application/octet-stream", rr.Header().Get("Content-Type"))
		mockStore.AssertExpectations(t)
	})

	t.Run("Not Found - 404 Not Found", func(t *testing.T) {
		// Arrange
		const notFoundUser = "not-found-user"
		mockStore := new(MockStore)
		mockStore.On("GetKey", notFoundUser).Return(nil, errors.New("not found"))

		apiHandler := &api.API{Store: mockStore, Logger: logger}
		req := httptest.NewRequest(http.MethodGet, "/keys/"+notFoundUser, nil)
		req.SetPathValue("userID", notFoundUser) // Set the path value for the test request
		rr := httptest.NewRecorder()

		// Act
		apiHandler.GetKeyHandler(rr, req)

		// Assert
		assert.Equal(t, http.StatusNotFound, rr.Code)
		mockStore.AssertExpectations(t)
	})
}
