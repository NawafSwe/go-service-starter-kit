package app

import (
	"fmt"

	grpcserver "github.com/nawafswe/go-service-starter-kit/internal/app/transport/grpc"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/clients/db/postgres"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/config"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/observability/tracing"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/worker"
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

	dbConn, err := postgres.NewConn(ctx, cfg.DB, fmt.Sprintf("%s.grpc.db", config.ServiceName), tp)
	if err != nil {
		lgr.Error(ctx, err, "[FATAL] failed to connect to database")
		return nil, err
	}

	srv, err := grpcserver.NewGRPCServer(ctx, cfg, dbConn, lgr)
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
	return withShutdown(grpcWorker, shutdown), nil
}
