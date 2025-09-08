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
)

// mockStore is a test double for the keyservice.Store interface.
type mockStore struct {
	StoreKeyFunc func(userID string, key []byte) error
	GetKeyFunc   func(userID string) ([]byte, error)
}

func (m *mockStore) StoreKey(userID string, key []byte) error {
	return m.StoreKeyFunc(userID, key)
}

func (m *mockStore) GetKey(userID string) ([]byte, error) {
	return m.GetKeyFunc(userID)
}

func TestKeyHandler(t *testing.T) {
	const testUserID = "user-123"
	const testKey = "my-public-key"
	logger := zerolog.Nop() // Use a no-op logger for all tests

	t.Run("POST /keys/{userID} - Success", func(t *testing.T) {
		// Arrange
		mock := &mockStore{
			StoreKeyFunc: func(userID string, key []byte) error {
				assert.Equal(t, testUserID, userID)
				assert.Equal(t, []byte(testKey), key)
				return nil
			},
		}
		apiHandler := &api.API{Store: mock, Logger: logger}
		req := httptest.NewRequest(http.MethodPost, "/keys/"+testUserID, bytes.NewReader([]byte(testKey)))
		rr := httptest.NewRecorder()

		// Act
		apiHandler.KeyHandler(rr, req)

		// Assert
		assert.Equal(t, http.StatusCreated, rr.Code)
	})

	t.Run("GET /keys/{userID} - Success", func(t *testing.T) {
		// Arrange
		mock := &mockStore{
			GetKeyFunc: func(userID string) ([]byte, error) {
				assert.Equal(t, testUserID, userID)
				return []byte(testKey), nil
			},
		}
		apiHandler := &api.API{Store: mock, Logger: logger}
		req := httptest.NewRequest(http.MethodGet, "/keys/"+testUserID, nil)
		rr := httptest.NewRecorder()

		// Act
		apiHandler.KeyHandler(rr, req)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, testKey, rr.Body.String())
		assert.Equal(t, "application/octet-stream", rr.Header().Get("Content-Type"))
	})

	t.Run("GET /keys/{userID} - Not Found", func(t *testing.T) {
		// Arrange
		mock := &mockStore{
			GetKeyFunc: func(userID string) ([]byte, error) {
				return nil, errors.New("not found")
			},
		}
		apiHandler := &api.API{Store: mock, Logger: logger}
		req := httptest.NewRequest(http.MethodGet, "/keys/not-found-user", nil)
		rr := httptest.NewRecorder()

		// Act
		apiHandler.KeyHandler(rr, req)

		// Assert
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("POST /keys/{userID} - Store Fails", func(t *testing.T) {
		// Arrange
		mock := &mockStore{
			StoreKeyFunc: func(userID string, key []byte) error {
				return errors.New("database is down")
			},
		}
		apiHandler := &api.API{Store: mock, Logger: logger}
		req := httptest.NewRequest(http.MethodPost, "/keys/"+testUserID, bytes.NewReader([]byte(testKey)))
		rr := httptest.NewRecorder()

		// Act
		apiHandler.KeyHandler(rr, req)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("Unsupported Method", func(t *testing.T) {
		// Arrange
		mock := &mockStore{}
		apiHandler := &api.API{Store: mock, Logger: logger}
		req := httptest.NewRequest(http.MethodDelete, "/keys/"+testUserID, nil)
		rr := httptest.NewRecorder()

		// Act
		apiHandler.KeyHandler(rr, req)

		// Assert
		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}
