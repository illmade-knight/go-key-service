// Package test provides public helpers for running end-to-end tests against the service.
package test

import (
	"net/http/httptest"

	"cloud.google.com/go/firestore"
	firestorestorage "github.com/illmade-knight/go-key-service/internal/storage/firestore"
	inmemorystore "github.com/illmade-knight/go-key-service/internal/storage/inmemory"
	"github.com/illmade-knight/go-key-service/keyservice"
	ks "github.com/illmade-knight/go-key-service/pkg/keyservice"
	"github.com/rs/zerolog"
)

// NewTestServer creates and starts a new httptest.Server for end-to-end testing.
// It assembles the service with an in-memory store.
func NewTestServer() *httptest.Server {
	cfg := &ks.Config{}
	store := inmemorystore.New()
	logger := zerolog.Nop()

	service := keyservice.New(cfg, store, logger)
	server := httptest.NewServer(service.Handler())

	return server
}

// NewTestKeyService creates and starts a new httptest.Server for the key service,
// backed by a real (emulated) Firestore client.
func NewTestKeyService(fsClient *firestore.Client, collectionName string) *httptest.Server {
	cfg := &ks.Config{}
	logger := zerolog.Nop()

	// This helper can legally import the internal storage package.
	store := firestorestorage.New(fsClient, collectionName)

	service := keyservice.New(cfg, store, logger)
	server := httptest.NewServer(service.Handler())

	return server
}
