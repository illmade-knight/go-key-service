package inmemory_test

import (
	"testing"

	"github.com/illmade-knight/go-key-service/internal/storage/inmemory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore(t *testing.T) {
	t.Run("GetKey after StoreKey returns the correct key", func(t *testing.T) {
		// Arrange
		store := inmemory.New()
		userID := "user-123"
		expectedKey := []byte("my-public-key")

		// Act
		err := store.StoreKey(userID, expectedKey)
		require.NoError(t, err)

		retrievedKey, err := store.GetKey(userID)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expectedKey, retrievedKey)
	})

	t.Run("GetKey for a non-existent user returns an error", func(t *testing.T) {
		// Arrange
		store := inmemory.New()

		// Act
		_, err := store.GetKey("non-existent-user")

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}
