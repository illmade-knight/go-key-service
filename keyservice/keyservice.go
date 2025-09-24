package keyservice

import (
	"net/http"

	"github.com/illmade-knight/go-key-service/internal/api"
	"github.com/illmade-knight/go-key-service/pkg/keyservice"
	"github.com/illmade-knight/go-microservice-base/pkg/microservice" // ADDED
	"github.com/rs/zerolog"
)

// Wrapper now embeds the BaseServer to inherit standard server functionality.
type Wrapper struct {
	*microservice.BaseServer // CHANGED: Embed BaseServer
	cfg                      *keyservice.Config
	logger                   zerolog.Logger
}

// New creates and wires up the entire key service.
func New(cfg *keyservice.Config, store keyservice.Store, logger zerolog.Logger) *Wrapper {
	// 1. Create the standard base server. It includes /healthz, /readyz, and /metrics.
	baseServer := microservice.NewBaseServer(logger, cfg.HTTPListenAddr)

	// 2. Create the service-specific API handlers.
	apiHandler := &api.API{Store: store, Logger: logger, JWTSecret: cfg.JWTSecret}

	// 3. Get the mux from the base server and register service-specific routes.
	mux := baseServer.Mux()

	storeKeyHandler := http.HandlerFunc(apiHandler.StoreKeyHandler)
	mux.Handle("POST /keys/{entityURN}", api.CorsMiddleware(apiHandler.JwtAuthMiddleware(storeKeyHandler)))

	doNothingHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	mux.Handle("OPTIONS /keys/{entityURN}", api.CorsMiddleware(doNothingHandler))
	mux.Handle("GET /keys/{entityURN}", api.CorsMiddleware(http.HandlerFunc(apiHandler.GetKeyHandler)))

	return &Wrapper{
		BaseServer: baseServer, // CHANGED
		cfg:        cfg,
		logger:     logger,
	}
}

// REMOVED: The Handler(), Start(), and Shutdown() methods are now inherited from BaseServer.
// If you needed to add custom shutdown logic (e.g., close a database client), you would
// override the Shutdown method here, perform your custom logic, and then call
// the embedded w.BaseServer.Shutdown(ctx).
