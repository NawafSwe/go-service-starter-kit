package bootstrap

import (
	"github.com/jmoiron/sqlx"
	"github.com/nawafswe/go-service-starter-kit/internal/app/repositories/example"
	"github.com/nawafswe/go-service-starter-kit/internal/observability/logger"
	"go.opentelemetry.io/otel/trace"
)

type (
	// Dependencies holds external client connections shared across the gRPC process.
	Dependencies struct {
		DBConn *sqlx.DB
	}

	// SharedResource holds cross-cutting concerns available to all handlers.
	SharedResource struct {
		Lgr    logger.Logger
		Tracer trace.Tracer
	}

	// SharedRepositories holds initialised repository instances.
	SharedRepositories struct {
		ExampleRepository example.Repository
	}
)
