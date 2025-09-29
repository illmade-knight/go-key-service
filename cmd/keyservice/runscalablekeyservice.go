package main

import (
	"context"
	"errors"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloud.google.com/go/firestore"
	fs "github.com/illmade-knight/go-key-service/internal/storage/firestore"
	"github.com/illmade-knight/go-key-service/keyservice"
	"github.com/illmade-knight/go-key-service/keyservice/config"
	ks "github.com/illmade-knight/go-key-service/pkg/keyservice"
	"github.com/illmade-knight/go-microservice-base/pkg/middleware"
	"github.com/rs/zerolog"
)

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	// --- 1. Load Configuration from YAML ---
	var configPath string
	flag.StringVar(&configPath, "config", "cmd/keyservice/local/local.yaml", "Path to config file")
	flag.Parse()

	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Override with environment variables for production secrets/settings
	if projectID := os.Getenv("GCP_PROJECT_ID"); projectID != "" {
		cfg.ProjectID = projectID
	}
	if idURL := os.Getenv("IDENTITY_SERVICE_URL"); idURL != "" {
		cfg.IdentityServiceURL = idURL
	}

	logger.Info().Str("run_mode", cfg.RunMode).Msg("Configuration loaded")

	// --- 2. Dependency Injection ---
	// REMOVED: The special-casing for "local" run_mode has been removed.
	// The service will now always connect to the real Firestore database.
	// In-memory fakes are correctly reserved for automated tests.
	fsClient, err := firestore.NewClient(context.Background(), cfg.ProjectID)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create Firestore client")
	}
	defer func() { _ = fsClient.Close() }()
	store := fs.New(fsClient, "public-keys")
	logger.Info().Str("project_id", cfg.ProjectID).Msg("Using Firestore key store")

	// --- 3. Service Initialization ---
	authMiddleware, err := middleware.NewJWKSAuthMiddleware(cfg.IdentityServiceURL)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create auth middleware")
	}

	// Create a ks.Config for the service New() function
	serviceCfg := &ks.Config{
		HTTPListenAddr: cfg.HTTPListenAddr,
		CorsConfig: middleware.CorsConfig{
			AllowedOrigins: cfg.Cors.AllowedOrigins,
			Role:           middleware.CorsRoleDefault,
		},
	}

	service := keyservice.New(serviceCfg, store, authMiddleware, logger)
	service.SetReady(true)

	// --- 4. Start Service and Handle Shutdown ---
	errChan := make(chan error, 1)
	go func() {
		logger.Info().Str("address", cfg.HTTPListenAddr).Msg("Starting service...")
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
		logger.Fatal().Err(err).Msg("Service shutdown failed")
	}

	logger.Info().Msg("Service stopped gracefully.")
}
