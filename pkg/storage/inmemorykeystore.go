package storage

import (
	"fmt"
	"sync"
)

// InMemoryStore is a thread-safe in-memory key store.
type InMemoryStore struct {
	sync.RWMutex
	keys map[string][]byte
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{keys: make(map[string][]byte)}
}

func (s *InMemoryStore) StoreKey(userID string, key []byte) error {
	s.Lock()
	defer s.Unlock()
	s.keys[userID] = key
	return nil
}

func (s *InMemoryStore) GetKey(userID string) ([]byte, error) {
	s.RLock()
	defer s.RUnlock()
	key, ok := s.keys[userID]
	if !ok {
		return nil, fmt.Errorf("key for user %s not found", userID)
	}
	return key, nil
}
