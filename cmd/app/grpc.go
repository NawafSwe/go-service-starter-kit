package app

import (
	"fmt"

	grpcserver "github.com/nawafswe/go-service-starter-kit/internal/app/transport/grpc"
	"github.com/nawafswe/go-service-starter-kit/internal/app/transport/grpc/bootstrap"
	"github.com/nawafswe/go-service-starter-kit/internal/clients/db/postgres"
	"github.com/nawafswe/go-service-starter-kit/internal/config"
	"github.com/nawafswe/go-service-starter-kit/internal/observability/tracing"
	"github.com/nawafswe/go-service-starter-kit/internal/worker"
)

// GRPCServerProcess wires up and starts the gRPC server.
type GRPCServerProcess struct{}

func NewGRPCServerProcess() GRPCServerProcess { return GRPCServerProcess{} }

func (g GRPCServerProcess) Register(args ProcessArgs) (Process, error) {
	ctx := args.Ctx
	cfg := args.Cfg
	lgr := args.Lgr

	lgr.Info(ctx, "initializing gRPC process...")

	tp, shutdown, err := tracing.Setup(ctx, cfg)
	if err != nil {
		lgr.Error(ctx, err, "[FATAL] failed to initialize tracer")
		return nil, err
	}

	// Build shared resources — tracer is optional.
	var resourceOpts []bootstrap.ResourceOption
	if cfg.General.Tracing.Enabled {
		resourceOpts = append(resourceOpts, bootstrap.WithTracer(tp.Tracer(worker.GRPCWorkerName)))
	}
	resources := bootstrap.NewSharedResource(lgr, resourceOpts...)

	var dbOpts []postgres.Option
	if cfg.General.Tracing.Enabled {
		dbOpts = append(dbOpts, postgres.WithTracerProvider(tp))
	}
	dbConn, err := postgres.NewConn(ctx, cfg.DB, fmt.Sprintf("%s.grpc.db", config.ServiceName), dbOpts...)
	if err != nil {
		lgr.Error(ctx, err, "[FATAL] failed to connect to database")
		return nil, err
	}

	deps := bootstrap.Dependencies{DBConn: dbConn}

	srv, err := grpcserver.NewGRPCServer(ctx, cfg, &deps, resources)
	if err != nil {
		lgr.Error(ctx, err, "[FATAL] failed to build gRPC server")
		return nil, err
	}
	lgr.Info(ctx, fmt.Sprintf("gRPC server listening on port %d", cfg.GRPC.Port))

	grpcWorker, err := worker.NewGRPCWorker(srv, lgr)
	if err != nil {
		lgr.Error(ctx, err, "[FATAL] failed to create gRPC worker")
		return nil, err
	}
	grpcWorker.WithAddr(fmt.Sprintf(":%d", cfg.GRPC.Port))
	return withShutdown(grpcWorker, shutdown), nil
}
