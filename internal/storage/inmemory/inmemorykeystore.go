// Package inmemory provides a thread-safe in-memory key store.
package inmemory

import (
	"fmt"
	"sync"
)

// Store is a concrete, thread-safe in-memory implementation of the keyservice.Store interface.
type Store struct {
	sync.RWMutex
	keys map[string][]byte
}

// New creates a new in-memory key store.
func New() *Store {
	return &Store{keys: make(map[string][]byte)}
}

// StoreKey adds a key to the in-memory map.
func (s *Store) StoreKey(userID string, key []byte) error {
	s.Lock()
	defer s.Unlock()
	s.keys[userID] = key
	return nil
}

// GetKey retrieves a key from the in-memory map.
func (s *Store) GetKey(userID string) ([]byte, error) {
	s.RLock()
	defer s.RUnlock()
	key, ok := s.keys[userID]
	if !ok {
		return nil, fmt.Errorf("key for user %s not found", userID)
	}
	return key, nil
}
