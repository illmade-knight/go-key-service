package main

import (
	"context"
	"errors"
	"log"
	"net/http" // ADDED
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloud.google.com/go/firestore"
	firestorestorage "github.com/illmade-knight/go-key-service/internal/storage/firestore"
	"github.com/illmade-knight/go-key-service/keyservice"
	ks "github.com/illmade-knight/go-key-service/pkg/keyservice"
	"github.com/rs/zerolog"
)

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	// --- 1. Configuration ---
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		logger.Fatal().Msg("JWT_SECRET environment variable must be set.")
	}
	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		logger.Fatal().Msg("GCP_PROJECT_ID environment variable must be set.")
	}
	cfg := &ks.Config{
		HTTPListenAddr: ":8081",
		JWTSecret:      jwtSecret,
	}

	// --- 2. Dependency Injection ---
	ctx := context.Background()
	fsClient, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create Firestore client")
	}
	defer fsClient.Close()
	store := firestorestorage.New(fsClient, "public-keys")
	logger.Info().Msg("Using Firestore key store")

	// --- 3. Service Initialization ---
	service := keyservice.New(cfg, store, logger)

	// ADDED: Signal that all dependencies are ready and the service can now pass readiness checks.
	service.SetReady(true)

	// --- 4. Start Service and Handle Shutdown (Standard Pattern) ---
	errChan := make(chan error, 1)
	go func() {
		logger.Info().Str("address", cfg.HTTPListenAddr).Msg("Starting service...")
		// The service's Start method is now blocking, inherited from BaseServer.
		if startErr := service.Start(); startErr != nil && !errors.Is(startErr, http.ErrServerClosed) {
			errChan <- startErr
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		logger.Fatal().Err(err).Msg("Service failed to start")
	case sig := <-quit:
		logger.Info().Str("signal", sig.String()).Msg("OS signal received, initiating shutdown.")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := service.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Service shutdown failed: %v", err)
	}

	logger.Info().Msg("Service stopped gracefully.")
}
