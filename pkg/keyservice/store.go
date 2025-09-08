// Package keyservice contains the public domain models, interfaces, and
// configuration for the key service. It defines the public contract.
package keyservice

// Store defines the public interface for key persistence.
// Any component that can store and retrieve keys (in-memory, Firestore, etc.)
// must implement this interface.
type Store interface {
	StoreKey(userID string, key []byte) error
	GetKey(userID string) ([]byte, error)
}
