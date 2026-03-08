package bootstrap

import (
	"github.com/nawafswe/go-service-starter-kit/internal/app/business/example"
	ep "github.com/nawafswe/go-service-starter-kit/internal/app/endpoint/v1"
	v1 "github.com/nawafswe/go-service-starter-kit/internal/app/transport/consumer/v1"
	"github.com/nawafswe/go-service-starter-kit/internal/config"
	kitconsumer "github.com/nawafswe/go-service-starter-kit/internal/gokit/consumer"
	"github.com/nawafswe/go-service-starter-kit/internal/middleware"
)

// InitializeRouter wires up the message router — one entry per consumed message type.
// Each entry pairs a transport-layer decode function with a go-kit endpoint.
func InitializeRouter(
	cfg config.Config,
	repos *SharedRepositories,
) MessageRouter {
	createHandler := example.NewCreateHandler(repos.ExampleRepository)

	return MessageRouter{
		"example.create": {
			Decode: v1.DecodeCreateExampleCommand,
			Endpoint: kitconsumer.MakeConsumerEndpoint(
				ep.MakeCreateExampleEndpoint(createHandler),
				middleware.TimeoutMiddleware(cfg.Endpoints.ExampleCreate.Deadline),
				middleware.RateLimit(cfg.Endpoints.ExampleCreate.RateLimiter.Interval, cfg.Endpoints.ExampleCreate.RateLimiter.Limit),
			),
		},
	}
}
