package bootstrap

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/auth"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/config"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/middleware"
)

// RegisterV1Routes registers all v1 API routes on the provided router.
//
// Route layout:
//   - /api/v1/examples           GET, POST  (auth required)
//   - /api/v1/examples/{id}      GET, DELETE (auth required)
func RegisterV1Routes(
	cfg config.Config,
	router *mux.Router,
	claimsParser auth.ClaimsParser,
	repos *SharedRepositories,
	resources *SharedResource,
) {
	lgr := resources.Lgr

	securedV1 := router.PathPrefix("/api/v1").Subrouter()
	securedV1.Use(middleware.AuthRequired(cfg.AuthorizationMock, claimsParser))

	securedV1.Handle("/examples", initializeCreateExampleHandler(cfg, lgr, repos)).Methods(http.MethodPost)
	securedV1.Handle("/examples", initializeListExamplesHandler(cfg, lgr, repos)).Methods(http.MethodGet)
	securedV1.Handle("/examples/{id}", initializeGetExampleHandler(cfg, lgr, repos)).Methods(http.MethodGet)
	securedV1.Handle("/examples/{id}", initializeDeleteExampleHandler(cfg, lgr, repos)).Methods(http.MethodDelete)
}
