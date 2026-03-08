package bootstrap

import (
	"github.com/nawafswe/go-service-starter-kit/internal/app/business/example"
	ep "github.com/nawafswe/go-service-starter-kit/internal/app/endpoint/v1"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/config"
	kitconsumer "github.com/nawafswe/go-service-starter-kit/internal/pkg/gokit/consumer"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/middleware"
)

// InitializeEndpoints wires up go-kit endpoints for each consumed message type.
func InitializeEndpoints(
	cfg config.Config,
	repos *SharedRepositories,
) *MessageEndpoints {
	createHandler := example.NewCreateHandler(repos.ExampleRepository)

	return &MessageEndpoints{
		CreateExample: kitconsumer.MakeConsumerEndpoint(
			ep.MakeCreateExampleEndpoint(createHandler),
			middleware.TimeoutMiddleware(cfg.Endpoints.ExampleCreate.Deadline),
			middleware.RateLimit(cfg.Endpoints.ExampleCreate.RateLimiter.Interval, cfg.Endpoints.ExampleCreate.RateLimiter.Limit),
		),
	}
}
