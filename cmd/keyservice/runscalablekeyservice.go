package main

import (
	"context"
	"log"
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

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		logger.Fatal().Msg("JWT_SECRET environment variable must be set.")
	}

	cfg := &ks.Config{
		HTTPListenAddr: ":8081",
		JWTSecret:      jwtSecret,
	}

	// 1. Create a real Firestore client for the production environment
	ctx := context.Background()
	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		logger.Fatal().Msg("GCP_PROJECT_ID environment variable must be set.")
	}
	fsClient, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create Firestore client")
	}
	defer fsClient.Close()

	// 2. Create the concrete Firestore storage implementation
	store := firestorestorage.New(fsClient, "public-keys")
	logger.Info().Msg("Using Firestore key store")

	// 3. Create the service wrapper, injecting the Firestore store
	service := keyservice.New(cfg, store, logger)

	// 4. Start the service and handle graceful shutdown
	go func() {
		if err := service.Start(); err != nil {
			logger.Fatal().Err(err).Msg("Failed to start key service")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("Shutdown signal received. Gracefully stopping service...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := service.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Service shutdown failed: %v", err)
	}

	logger.Info().Msg("Service stopped gracefully.")
}
