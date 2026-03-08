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

	// MessageHandler pairs a decode function with a go-kit endpoint.
	// Decode converts the raw payload into the request type the endpoint expects.
	MessageHandler struct {
		Decode   func(payload []byte) (any, error)
		Endpoint endpoint.Endpoint
	}

	// MessageRouter maps message type names to their handler.
	// Used by the consumer to route incoming messages.
	MessageRouter map[string]MessageHandler
)
