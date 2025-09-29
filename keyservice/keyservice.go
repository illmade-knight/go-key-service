package keyservice

import (
	"net/http"

	"github.com/illmade-knight/go-key-service/internal/api"
	"github.com/illmade-knight/go-key-service/pkg/keyservice"
	"github.com/illmade-knight/go-microservice-base/pkg/microservice"
	"github.com/illmade-knight/go-microservice-base/pkg/middleware"
	"github.com/rs/zerolog"
)

// Wrapper now embeds the BaseServer to inherit standard server functionality.
type Wrapper struct {
	*microservice.BaseServer
	logger zerolog.Logger
}

// New creates and wires up the entire key service.
func New(
	cfg *keyservice.Config,
	store keyservice.Store,
	authMiddleware func(http.Handler) http.Handler, // Accept middleware via DI
	logger zerolog.Logger,
) *Wrapper {
	// 1. Create the standard base server.
	baseServer := microservice.NewBaseServer(logger, cfg.HTTPListenAddr)

	// 2. Create the service-specific API handlers.
	apiHandler := &api.API{Store: store, Logger: logger}

	// 3. Get the mux from the base server and register routes.
	mux := baseServer.Mux()

	// 4. Create CORS middleware from the loaded config.
	corsMiddleware := middleware.NewCorsMiddleware(middleware.CorsConfig{
		AllowedOrigins: cfg.CorsConfig.AllowedOrigins,
		Role:           middleware.CorsRoleDefault,
	})

	// 5. Apply middleware to the handlers.
	storeKeyHandler := http.HandlerFunc(apiHandler.StoreKeyHandler)
	mux.Handle("POST /keys/{entityURN}", corsMiddleware(authMiddleware(storeKeyHandler)))

	// Public endpoint, only needs CORS.
	getKeyHandler := http.HandlerFunc(apiHandler.GetKeyHandler)
	mux.Handle("GET /keys/{entityURN}", corsMiddleware(getKeyHandler))

	// OPTIONS handler for CORS preflight requests.
	optionsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	mux.Handle("OPTIONS /keys/{entityURN}", corsMiddleware(optionsHandler))

	return &Wrapper{
		BaseServer: baseServer,
		logger:     logger,
	}
}
