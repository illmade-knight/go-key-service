// REFACTOR: This test file is updated to validate the URN-based and
// context-aware in-memory store implementation.

package inmemory_test

import (
	"context"
	"testing"

	"github.com/illmade-knight/go-key-service/internal/storage/inmemory"
	"github.com/illmade-knight/go-secure-messaging/pkg/urn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore(t *testing.T) {
	ctx := context.Background()

	t.Run("GetKey after StoreKey returns the correct key", func(t *testing.T) {
		// Arrange
		store := inmemory.New()
		testURN, err := urn.New("user", "user-123", urn.SecureMessaging)
		require.NoError(t, err)
		expectedKey := []byte("my-public-key")

		// Act
		err = store.StoreKey(ctx, testURN, expectedKey)
		require.NoError(t, err)

		retrievedKey, err := store.GetKey(ctx, testURN)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expectedKey, retrievedKey)
	})

	t.Run("GetKey for a non-existent entity returns an error", func(t *testing.T) {
		// Arrange
		store := inmemory.New()
		nonExistentURN, err := urn.New("user", "non-existent-user", urn.SecureMessaging)
		require.NoError(t, err)

		// Act
		_, err = store.GetKey(ctx, nonExistentURN)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}
