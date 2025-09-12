// REFACTOR: This file is updated to implement the new URN-based and
// context-aware keyservice.Store interface. All Firestore operations now use
// the canonical string representation of the URN as the document key.

// Package firestore provides a key store implementation using Google Cloud Firestore.
package firestore

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/illmade-knight/go-secure-messaging/pkg/urn"
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

// StoreKey creates or overwrites a document with the entity's public key.
func (s *Store) StoreKey(ctx context.Context, entityURN urn.URN, key []byte) error {
	entityKey := entityURN.String()
	doc := s.collection.Doc(entityKey)
	_, err := doc.Set(ctx, keyDocument{PublicKey: key})
	if err != nil {
		return fmt.Errorf("failed to store key for entity %s: %w", entityKey, err)
	}
	return nil
}

// GetKey retrieves an entity's public key from a Firestore document.
func (s *Store) GetKey(ctx context.Context, entityURN urn.URN) ([]byte, error) {
	entityKey := entityURN.String()
	doc, err := s.collection.Doc(entityKey).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, fmt.Errorf("key for entity %s not found", entityKey)
		}
		return nil, fmt.Errorf("failed to get key for entity %s: %w", entityKey, err)
	}

	var kd keyDocument
	if err := doc.DataTo(&kd); err != nil {
		return nil, fmt.Errorf("failed to decode key for entity %s: %w", entityKey, err)
	}
	return kd.PublicKey, nil
}
