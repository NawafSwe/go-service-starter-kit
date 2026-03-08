package app

import (
	"fmt"

	httpserver "github.com/nawafswe/go-service-starter-kit/internal/app/transport/http"
	"github.com/nawafswe/go-service-starter-kit/internal/app/transport/http/bootstrap"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/clients/db/postgres"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/config"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/observability/tracing"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/worker"
)

// HTTPServerProcess wires up and starts the HTTP server.
type HTTPServerProcess struct{}

func NewHTTPServerProcess() HTTPServerProcess { return HTTPServerProcess{} }

func (h HTTPServerProcess) Register(args ProcessArgs) (Process, error) {
	ctx := args.Ctx
	cfg := args.Cfg
	lgr := args.Lgr

	lgr.Info(ctx, "initializing HTTP process...")

	tp, shutdown, err := tracing.Setup(ctx, cfg)
	if err != nil {
		lgr.Error(ctx, err, "[FATAL] failed to initialize tracer")
		return nil, err
	}

	resources := bootstrap.SharedResource{
		Lgr:            lgr,
		Tracer:         tp.Tracer(worker.HTTPWorkerName),
		MetricProvider: args.MeterProvider,
		Meter:          args.Meter,
	}

	dbConn, err := postgres.NewConn(ctx, cfg.DB, fmt.Sprintf("%s.db", config.ServiceName), tp)
	if err != nil {
		lgr.Error(ctx, err, "[FATAL] failed to connect to database")
		return nil, err
	}

	deps, err := bootstrap.InitializeClients(cfg, dbConn, &resources)
	if err != nil {
		lgr.Error(ctx, err, "[FATAL] failed to initialize dependencies")
		return nil, err
	}

	srv, err := httpserver.NewHTTPServer(ctx, cfg, &deps, &resources)
	if err != nil {
		lgr.Error(ctx, err, "[FATAL] failed to build HTTP server")
		return nil, err
	}
	lgr.Info(ctx, fmt.Sprintf("HTTP server listening on port %d", cfg.HTTP.Port))

	httpWorker, err := worker.NewHTTPWorker(srv, lgr)
	if err != nil {
		lgr.Error(ctx, err, "[FATAL] failed to create HTTP worker")
		return nil, err
	}
	return withShutdown(httpWorker, shutdown), nil
}
