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
		Tracer trace.Tracer // optional — nil means no tracing
	}

	// SharedRepositories holds initialised repository instances.
	SharedRepositories struct {
		ExampleRepository example.Repository
	}
)

// ResourceOption configures optional fields on SharedResource.
type ResourceOption func(*SharedResource)

// NewSharedResource creates a SharedResource with the required logger and optional
// observability concerns. Without any options, tracing is disabled.
func NewSharedResource(lgr logger.Logger, opts ...ResourceOption) *SharedResource {
	r := &SharedResource{Lgr: lgr}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// WithTracer enables distributed tracing on the gRPC server.
func WithTracer(t trace.Tracer) ResourceOption {
	return func(r *SharedResource) { r.Tracer = t }
}
