// Package firestore provides a key store implementation using Google Cloud Firestore.
package firestore

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// keyDocument is the structure stored in a Firestore document.
type keyDocument struct {
	PublicKey []byte `firestore:"publicKey"`
}

// Store is a concrete implementation of the keyservice.Store interface using Firestore.
type Store struct {
	client     *firestore.Client
	collection *firestore.CollectionRef
}

// New creates a new Firestore-backed store.
func New(client *firestore.Client, collectionName string) *Store {
	return &Store{
		client:     client,
		collection: client.Collection(collectionName),
	}
}

// StoreKey creates or overwrites a document with the user's public key.
func (s *Store) StoreKey(userID string, key []byte) error {
	doc := s.collection.Doc(userID)
	_, err := doc.Set(context.Background(), keyDocument{PublicKey: key})
	if err != nil {
		return fmt.Errorf("failed to store key for user %s: %w", userID, err)
	}
	return nil
}

// GetKey retrieves a user's public key from a Firestore document.
func (s *Store) GetKey(userID string) ([]byte, error) {
	doc, err := s.collection.Doc(userID).Get(context.Background())
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, fmt.Errorf("key for user %s not found", userID)
		}
		return nil, fmt.Errorf("failed to get key for user %s: %w", userID, err)
	}

	var kd keyDocument
	if err := doc.DataTo(&kd); err != nil {
		return nil, fmt.Errorf("failed to decode key for user %s: %w", userID, err)
	}
	return kd.PublicKey, nil
}
