package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/illmade-knight/go-key-service/internal/storage/inmemory"
	"github.com/illmade-knight/go-key-service/keyservice"
	ks "github.com/illmade-knight/go-key-service/pkg/keyservice"
	"github.com/rs/zerolog"
)

// api holds the dependencies for our service, like the storage layer.
type api struct {
	store *inmemory.Store
}

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	
	cfg := &ks.Config{
		HTTPListenAddr: ":8081",
	}

	// 1. Create the inmemory storage implementation
	store := inmemory.New()
	logger.Info().Msg("Using Inmemory key store")

	// 2. Create the service wrapper, injecting the inmemory store
	service := keyservice.New(cfg, store, logger)

	// 3. Start the service and handle graceful shutdown
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
