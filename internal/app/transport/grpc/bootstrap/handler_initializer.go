package bootstrap

import (
	grpctransport "github.com/go-kit/kit/transport/grpc"

	"github.com/nawafswe/go-service-starter-kit/internal/app/business/example"
	ep "github.com/nawafswe/go-service-starter-kit/internal/app/endpoint/v1"
	v1 "github.com/nawafswe/go-service-starter-kit/internal/app/transport/grpc/v1"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/config"
	kitgrpc "github.com/nawafswe/go-service-starter-kit/internal/pkg/gokit/grpc"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/middleware"
)

func initializeCreateExampleHandler(
	cfg config.Config,
	resources *SharedResource,
	repos *SharedRepositories,
) grpctransport.Handler {
	handler := example.NewCreateHandler(repos.ExampleRepository)
	return kitgrpc.MakeGRPCHandler(
		ep.MakeCreateExampleEndpoint(handler),
		v1.DecodeCreateExampleRequest,
		v1.EncodeCreateExampleResponse,
		resources.Lgr,
		middleware.TimeoutMiddleware(cfg.Endpoints.ExampleCreate.Deadline),
		middleware.RateLimit(cfg.Endpoints.ExampleCreate.RateLimiter.Interval, cfg.Endpoints.ExampleCreate.RateLimiter.Limit),
	)
}

func initializeGetExampleHandler(
	cfg config.Config,
	resources *SharedResource,
	repos *SharedRepositories,
) grpctransport.Handler {
	handler := example.NewGetHandler(repos.ExampleRepository)
	return kitgrpc.MakeGRPCHandler(
		ep.MakeGetExampleEndpoint(handler),
		v1.DecodeGetExampleRequest,
		v1.EncodeGetExampleResponse,
		resources.Lgr,
		middleware.TimeoutMiddleware(cfg.Endpoints.ExampleGet.Deadline),
		middleware.RateLimit(cfg.Endpoints.ExampleGet.RateLimiter.Interval, cfg.Endpoints.ExampleGet.RateLimiter.Limit),
	)
}

func initializeListExamplesHandler(
	cfg config.Config,
	resources *SharedResource,
	repos *SharedRepositories,
) grpctransport.Handler {
	handler := example.NewListHandler(repos.ExampleRepository)
	return kitgrpc.MakeGRPCHandler(
		ep.MakeListExamplesEndpoint(handler),
		v1.DecodeListExamplesRequest,
		v1.EncodeListExamplesResponse,
		resources.Lgr,
		middleware.TimeoutMiddleware(cfg.Endpoints.ExampleList.Deadline),
		middleware.RateLimit(cfg.Endpoints.ExampleList.RateLimiter.Interval, cfg.Endpoints.ExampleList.RateLimiter.Limit),
	)
}

func initializeDeleteExampleHandler(
	cfg config.Config,
	resources *SharedResource,
	repos *SharedRepositories,
) grpctransport.Handler {
	handler := example.NewDeleteHandler(repos.ExampleRepository)
	return kitgrpc.MakeGRPCHandler(
		ep.MakeDeleteExampleEndpoint(handler),
		v1.DecodeDeleteExampleRequest,
		v1.EncodeDeleteExampleResponse,
		resources.Lgr,
		middleware.TimeoutMiddleware(cfg.Endpoints.ExampleDelete.Deadline),
		middleware.RateLimit(cfg.Endpoints.ExampleDelete.RateLimiter.Interval, cfg.Endpoints.ExampleDelete.RateLimiter.Limit),
	)
}

// InitializeServiceHandler wires up and returns the ExampleServiceHandler.
func InitializeServiceHandler(
	cfg config.Config,
	resources *SharedResource,
	repos *SharedRepositories,
) *v1.ExampleServiceHandler {
	return v1.NewExampleServiceHandler(
		initializeCreateExampleHandler(cfg, resources, repos),
		initializeGetExampleHandler(cfg, resources, repos),
		initializeListExamplesHandler(cfg, resources, repos),
		initializeDeleteExampleHandler(cfg, resources, repos),
	)
}
