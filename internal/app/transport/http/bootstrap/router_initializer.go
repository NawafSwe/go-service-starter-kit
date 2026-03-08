package bootstrap

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nawafswe/go-service-starter-kit/internal/auth"
	"github.com/nawafswe/go-service-starter-kit/internal/config"
	"github.com/nawafswe/go-service-starter-kit/internal/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// InitializeRoutes builds and returns the root mux.Router.
func InitializeRoutes(
	ctx context.Context,
	cfg config.Config,
	repos *SharedRepositories,
	deps *Dependencies,
	resources *SharedResource,
) (*mux.Router, error) {
	router := mux.NewRouter()
	router.Use(mux.CORSMethodMiddleware(router))
	if resources.Tracer != nil {
		router.Use(otelhttp.NewMiddleware("http.server"))
	}

	// Health check — wires DB connectivity into the liveness probe.
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := deps.DBConn.PingContext(ctx); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("unhealthy"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("healthy"))
	}).Methods(http.MethodGet)

	claimsParser, err := auth.NewClaimsParser(cfg.JWT.ISSUER, []byte(cfg.JWT.Secret))
	if err != nil {
		return nil, fmt.Errorf("failed to create claims parser: %w", err)
	}

	// AuthOptional at the global level so public endpoints still receive user context
	// when a valid token is present.
	router.Use(middleware.AuthOptional(cfg.AuthorizationMock, claimsParser))

	appRouter := router.PathPrefix("/app").Subrouter()
	RegisterV1Routes(cfg, appRouter, claimsParser, repos, resources)

	return router, nil
}
