package storage

// Store defines the interface for key persistence.
type Store interface {
	StoreKey(userID string, key []byte) error
	GetKey(userID string) ([]byte, error)
}
