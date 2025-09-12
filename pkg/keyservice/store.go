// REFACTOR: This file is updated to use the canonical urn.URN type from the
// go-secure-messaging library. This makes the key service a generic store for
// any entity in our system. Context has also been added to the interface
// methods to align with best practices.

// Package keyservice contains the public domain models, interfaces, and
// configuration for the key service. It defines the public contract.
package keyservice

import (
	"context"

	"github.com/illmade-knight/go-secure-messaging/pkg/urn"
)

// Store defines the public interface for key persistence.
// Any component that can store and retrieve keys (in-memory, Firestore, etc.)
// must implement this interface.
type Store interface {
	StoreKey(ctx context.Context, entityURN urn.URN, key []byte) error
	GetKey(ctx context.Context, entityURN urn.URN) ([]byte, error)
}
