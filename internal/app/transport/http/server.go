package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/nawafswe/go-service-starter-kit/internal/app/transport/http/bootstrap"
	"github.com/nawafswe/go-service-starter-kit/internal/config"
)

// NewHTTPServer builds the HTTP server — initialises repositories, wires routes,
// and returns a configured http.Server ready to be handed to the HTTP worker.
func NewHTTPServer(
	ctx context.Context,
	cfg config.Config,
	deps *bootstrap.Dependencies,
	resources *bootstrap.SharedResource,
) (*http.Server, error) {
	repos, err := bootstrap.InitializeRepositories(cfg, deps, resources)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize repositories: %w", err)
	}
	router, err := bootstrap.InitializeRoutes(ctx, cfg, repos, deps, resources)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize routes: %w", err)
	}
	return &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.HTTP.Port),
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout,
		Handler:           router,
	}, nil
}
