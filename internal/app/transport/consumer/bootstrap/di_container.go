package bootstrap

import (
	"github.com/go-kit/kit/endpoint"
	"github.com/jmoiron/sqlx"
	"github.com/nawafswe/go-service-starter-kit/internal/app/repositories/example"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/observability/logger"
)

type (
	// Dependencies holds external client connections shared across the consumer process.
	Dependencies struct {
		DBConn *sqlx.DB
	}

	// SharedResource holds cross-cutting concerns available to all handlers.
	SharedResource struct {
		Lgr logger.Logger
	}

	// SharedRepositories holds initialised repository instances.
	SharedRepositories struct {
		ExampleRepository example.Repository
	}

	// MessageEndpoints holds go-kit endpoints for each consumed message type.
	// Each endpoint has the full middleware chain (timeout, rate-limit, logging).
	MessageEndpoints struct {
		CreateExample endpoint.Endpoint
	}
)
