// context-aware keyservice.Store interface.

// Package inmemory provides a thread-safe in-memory key store.
package inmemory

import (
	"context"
	"fmt"
	"sync"

	"github.com/illmade-knight/go-secure-messaging/pkg/urn"
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

// StoreKey adds a key to the in-memory map using the URN's string representation as the key.
func (s *Store) StoreKey(ctx context.Context, entityURN urn.URN, key []byte) error {
	s.Lock()
	defer s.Unlock()
	s.keys[entityURN.String()] = key
	return nil
}

// GetKey retrieves a key from the in-memory map using the URN's string representation.
func (s *Store) GetKey(ctx context.Context, entityURN urn.URN) ([]byte, error) {
	s.RLock()
	defer s.RUnlock()
	key, ok := s.keys[entityURN.String()]
	if !ok {
		return nil, fmt.Errorf("key for entity %s not found", entityURN.String())
	}
	return key, nil
}
