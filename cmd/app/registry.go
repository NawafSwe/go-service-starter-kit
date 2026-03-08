package app

import (
	"context"

	"github.com/nawafswe/go-service-starter-kit/internal/config"
	"github.com/nawafswe/go-service-starter-kit/internal/observability/logger"
	otelmeter "go.opentelemetry.io/otel/metric"
)

// ProcessArgs encapsulates shared dependencies passed to every process at startup.
type ProcessArgs struct {
	Ctx           context.Context
	Cfg           config.Config
	Lgr           logger.Logger
	MeterProvider otelmeter.MeterProvider
	Meter         otelmeter.Meter
}

// Process is the interface that every long-running process must implement.
type Process interface {
	Run(ctx context.Context) error
}

// ProcessRegistry wires up and returns a runnable Process.
type ProcessRegistry interface {
	Register(args ProcessArgs) (Process, error)
}

// RegistryProcessesMap registers all long-running processes by name.
// Add new processes (http, grpc, consumer, …) here.
var RegistryProcessesMap = map[string]ProcessRegistry{
	"http":     NewHTTPServerProcess(),
	"grpc":     NewGRPCServerProcess(),
	"consumer": NewConsumerProcess(),
}

// withShutdown wraps a Process so that shutdown is called after Run returns,
// ensuring trace/log exporters are flushed even on error.
func withShutdown(p Process, shutdown func(context.Context) error) Process {
	return &processWithShutdown{inner: p, shutdown: shutdown}
}

type processWithShutdown struct {
	inner    Process
	shutdown func(context.Context) error
}

func (p *processWithShutdown) Run(ctx context.Context) error {
	defer p.shutdown(ctx) //nolint:errcheck
	return p.inner.Run(ctx)
}
