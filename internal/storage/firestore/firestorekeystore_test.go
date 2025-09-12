// REFACTOR: This new test file provides local unit tests for the Firestore
// store, ensuring it correctly implements the URN-based interface.

package firestore_test

import (
	"context"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	fsAdaper "github.com/illmade-knight/go-key-service/internal/storage/firestore"
	"github.com/illmade-knight/go-key-service/pkg/keyservice"
	"github.com/illmade-knight/go-secure-messaging/pkg/urn"
	"github.com/illmade-knight/go-test/emulators"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupSuite initializes a Firestore emulator and a new Store for testing.
func setupSuite(t *testing.T) (context.Context, *firestore.Client, keyservice.Store) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(cancel)

	const projectID = "test-project-keystore"
	const collectionName = "public-keys"

	firestoreConn := emulators.SetupFirestoreEmulator(t, ctx, emulators.GetDefaultFirestoreConfig(projectID))
	fsClient, err := firestore.NewClient(context.Background(), projectID, firestoreConn.ClientOptions...)
	require.NoError(t, err)
	t.Cleanup(func() { _ = fsClient.Close() })

	store := fsAdaper.New(fsClient, collectionName)

	return ctx, fsClient, store
}

func TestFirestoreStore_Integration(t *testing.T) {
	ctx, _, store := setupSuite(t)

	// Arrange
	userURN, err := urn.New("user", "user-123", urn.SecureMessaging)
	require.NoError(t, err)
	deviceURN, err := urn.New("device", "device-abc", urn.SecureMessaging)
	require.NoError(t, err)

	userKey := []byte("user-public-key")
	deviceKey := []byte("device-public-key")

	// Act & Assert: Store and retrieve a user key
	err = store.StoreKey(ctx, userURN, userKey)
	require.NoError(t, err)

	retrievedUserKey, err := store.GetKey(ctx, userURN)
	require.NoError(t, err)
	assert.Equal(t, userKey, retrievedUserKey)

	// Act & Assert: Store and retrieve a device key
	err = store.StoreKey(ctx, deviceURN, deviceKey)
	require.NoError(t, err)

	retrievedDeviceKey, err := store.GetKey(ctx, deviceURN)
	require.NoError(t, err)
	assert.Equal(t, deviceKey, retrievedDeviceKey)

	// Act & Assert: Getting a non-existent key returns an error
	nonExistentURN, err := urn.New("user", "not-real", urn.SecureMessaging)
	require.NoError(t, err)

	_, err = store.GetKey(ctx, nonExistentURN)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
