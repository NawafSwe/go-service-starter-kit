package bootstrap

import (
	"github.com/jmoiron/sqlx"
	"github.com/nawafswe/go-service-starter-kit/internal/app/repositories/example"
	"github.com/nawafswe/go-service-starter-kit/internal/observability/logger"
	otelmeter "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type (
	// Dependencies holds external client connections shared across the HTTP process.
	Dependencies struct {
		DBConn *sqlx.DB
	}

	// SharedResource holds cross-cutting concerns available to all handlers.
	SharedResource struct {
		Lgr            logger.Logger
		Tracer         trace.Tracer
		MetricProvider otelmeter.MeterProvider
		Meter          otelmeter.Meter
	}

	// SharedRepositories holds initialised repository instances.
	SharedRepositories struct {
		ExampleRepository example.Repository
	}
)
