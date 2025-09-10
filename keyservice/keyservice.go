// Package keyservice provides the main, embeddable service wrapper.
package keyservice

import (
	"context"
	"errors"
	"net/http"

	"github.com/illmade-knight/go-key-service/internal/api"
	"github.com/illmade-knight/go-key-service/pkg/keyservice"
	"github.com/rs/zerolog"
)

// Wrapper encapsulates all components of the running service.
type Wrapper struct {
	cfg    *keyservice.Config
	server *http.Server
	logger zerolog.Logger
}

// New creates and wires up the entire key service.
func New(cfg *keyservice.Config, store keyservice.Store, logger zerolog.Logger) *Wrapper {
	apiHandler := &api.API{Store: store, Logger: logger}

	mux := http.NewServeMux()

	// This handler does nothing, but it's needed to complete the middleware chain for OPTIONS.
	doNothingHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	// ADD THIS LINE to handle the browser's OPTIONS preflight request.
	mux.Handle("OPTIONS /keys/{userID}", api.CorsMiddleware(doNothingHandler))
	mux.Handle("POST /keys/{userID}", api.CorsMiddleware(http.HandlerFunc(apiHandler.StoreKeyHandler))) // Use Handle instead of HandleFunc
	mux.Handle("GET /keys/{userID}", api.CorsMiddleware(http.HandlerFunc(apiHandler.GetKeyHandler)))    // Use Handle instead of HandleFunc

	return &Wrapper{
		cfg:    cfg,
		server: &http.Server{Addr: cfg.HTTPListenAddr, Handler: mux},
		logger: logger,
	}
}

// Handler returns the underlying http.Handler for the service.
func (w *Wrapper) Handler() http.Handler {
	return w.server.Handler
}

// Start runs the service's HTTP server.
func (w *Wrapper) Start() error {
	w.logger.Info().Str("address", w.cfg.HTTPListenAddr).Msg("Key Service starting...")
	if err := w.server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Shutdown gracefully stops the service.
func (w *Wrapper) Shutdown(ctx context.Context) error {
	w.logger.Info().Msg("Key Service shutting down...")
	return w.server.Shutdown(ctx)
}
