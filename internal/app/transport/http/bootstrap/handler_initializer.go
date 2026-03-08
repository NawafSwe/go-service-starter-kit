package bootstrap

import (
	"net/http"

	"github.com/nawafswe/go-service-starter-kit/internal/app/business/example"
	ep "github.com/nawafswe/go-service-starter-kit/internal/app/endpoint/v1"
	"github.com/nawafswe/go-service-starter-kit/internal/app/transport/http/encoder"
	v1 "github.com/nawafswe/go-service-starter-kit/internal/app/transport/http/v1"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/config"
	kithttp "github.com/nawafswe/go-service-starter-kit/internal/pkg/gokit/http"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/middleware"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/observability/logger"
)

func initializeCreateExampleHandler(
	cfg config.Config,
	lgr logger.Logger,
	repos *SharedRepositories,
) http.Handler {
	handler := example.NewCreateHandler(repos.ExampleRepository)
	codec := v1.CreateExampleEncoderDecoder{}
	return kithttp.MakeHTTPHandler(
		ep.MakeCreateExampleEndpoint(handler),
		codec,
		codec,
		encoder.NewJSONErrorEncoder(),
		lgr,
		middleware.TimeoutMiddleware(cfg.Endpoints.ExampleCreate.Deadline),
		middleware.RateLimit(cfg.Endpoints.ExampleCreate.RateLimiter.Interval, cfg.Endpoints.ExampleCreate.RateLimiter.Limit),
	)
}

func initializeGetExampleHandler(
	cfg config.Config,
	lgr logger.Logger,
	repos *SharedRepositories,
) http.Handler {
	handler := example.NewGetHandler(repos.ExampleRepository)
	codec := v1.GetExampleEncoderDecoder{}
	return kithttp.MakeHTTPHandler(
		ep.MakeGetExampleEndpoint(handler),
		codec,
		codec,
		encoder.NewJSONErrorEncoder(),
		lgr,
		middleware.TimeoutMiddleware(cfg.Endpoints.ExampleGet.Deadline),
		middleware.RateLimit(cfg.Endpoints.ExampleGet.RateLimiter.Interval, cfg.Endpoints.ExampleGet.RateLimiter.Limit),
	)
}

func initializeListExamplesHandler(
	cfg config.Config,
	lgr logger.Logger,
	repos *SharedRepositories,
) http.Handler {
	handler := example.NewListHandler(repos.ExampleRepository)
	codec := v1.ListExamplesEncoderDecoder{}
	return kithttp.MakeHTTPHandler(
		ep.MakeListExamplesEndpoint(handler),
		codec,
		codec,
		encoder.NewJSONErrorEncoder(),
		lgr,
		middleware.TimeoutMiddleware(cfg.Endpoints.ExampleList.Deadline),
		middleware.RateLimit(cfg.Endpoints.ExampleList.RateLimiter.Interval, cfg.Endpoints.ExampleList.RateLimiter.Limit),
	)
}

func initializeDeleteExampleHandler(
	cfg config.Config,
	lgr logger.Logger,
	repos *SharedRepositories,
) http.Handler {
	handler := example.NewDeleteHandler(repos.ExampleRepository)
	codec := v1.DeleteExampleEncoderDecoder{}
	return kithttp.MakeHTTPHandler(
		ep.MakeDeleteExampleEndpoint(handler),
		codec,
		codec,
		encoder.NewJSONErrorEncoder(),
		lgr,
		middleware.TimeoutMiddleware(cfg.Endpoints.ExampleDelete.Deadline),
		middleware.RateLimit(cfg.Endpoints.ExampleDelete.RateLimiter.Interval, cfg.Endpoints.ExampleDelete.RateLimiter.Limit),
	)
}
