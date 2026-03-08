package worker

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/nawafswe/go-service-starter-kit/internal/pkg/observability/logger"
	"google.golang.org/grpc"
)

const GRPCWorkerName = "grpc-worker"

// GRPCWorker runs a grpc.Server with graceful shutdown on SIGINT/SIGTERM.
type GRPCWorker struct {
	srv  *grpc.Server
	addr string
	lgr  logger.Logger
}

func NewGRPCWorker(srv *grpc.Server, lgr logger.Logger) (*GRPCWorker, error) {
	if srv == nil {
		return nil, fmt.Errorf("server is required to create %s", GRPCWorkerName)
	}
	if lgr == nil {
		return nil, fmt.Errorf("logger is required to create %s", GRPCWorkerName)
	}
	return &GRPCWorker{srv: srv, lgr: lgr}, nil
}

// WithAddr sets the listen address. Call before Run.
func (g *GRPCWorker) WithAddr(addr string) *GRPCWorker {
	g.addr = addr
	return g
}

func (g *GRPCWorker) Run(ctx context.Context) error {
	addr := g.addr
	if addr == "" {
		addr = ":50051"
	}

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("gRPC worker: listen on %s: %w", addr, err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(stop)

	errCh := make(chan error, 1)

	g.lgr.Info(ctx, fmt.Sprintf("gRPC server listening on %s", addr))
	go func() {
		if err := g.srv.Serve(lis); err != nil {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case sig := <-stop:
		g.lgr.Info(ctx, fmt.Sprintf("shutdown signal received: %s", sig))
	case err := <-errCh:
		if err != nil {
			g.lgr.Error(ctx, err, "gRPC server stopped with error")
			return err
		}
	}

	g.lgr.Info(ctx, "shutting down gRPC server...")
	g.srv.GracefulStop()
	g.lgr.Info(ctx, "gRPC server shutdown complete")
	return nil
}
