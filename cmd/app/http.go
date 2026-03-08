package app

import (
	"fmt"

	httpserver "github.com/nawafswe/go-service-starter-kit/internal/app/transport/http"
	"github.com/nawafswe/go-service-starter-kit/internal/app/transport/http/bootstrap"
	"github.com/nawafswe/go-service-starter-kit/internal/clients/db/postgres"
	"github.com/nawafswe/go-service-starter-kit/internal/config"
	"github.com/nawafswe/go-service-starter-kit/internal/observability/tracing"
	"github.com/nawafswe/go-service-starter-kit/internal/worker"
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

	// Build shared resources — tracer and metrics are optional.
	var resourceOpts []bootstrap.ResourceOption
	if cfg.General.Tracing.Enabled {
		resourceOpts = append(resourceOpts, bootstrap.WithTracer(tp.Tracer(worker.HTTPWorkerName)))
	}
	if args.MeterProvider != nil {
		resourceOpts = append(resourceOpts, bootstrap.WithMeterProvider(args.MeterProvider))
	}
	if args.Meter != nil {
		resourceOpts = append(resourceOpts, bootstrap.WithMeter(args.Meter))
	}
	resources := bootstrap.NewSharedResource(lgr, resourceOpts...)

	var dbOpts []postgres.Option
	if cfg.General.Tracing.Enabled {
		dbOpts = append(dbOpts, postgres.WithTracerProvider(tp))
	}
	dbConn, err := postgres.NewConn(ctx, cfg.DB, fmt.Sprintf("%s.db", config.ServiceName), dbOpts...)
	if err != nil {
		lgr.Error(ctx, err, "[FATAL] failed to connect to database")
		return nil, err
	}

	deps, err := bootstrap.InitializeClients(cfg, dbConn, resources)
	if err != nil {
		lgr.Error(ctx, err, "[FATAL] failed to initialize dependencies")
		return nil, err
	}

	srv, err := httpserver.NewHTTPServer(ctx, cfg, &deps, resources)
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
