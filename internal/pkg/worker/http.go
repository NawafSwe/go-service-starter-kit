package worker

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nawafswe/go-service-starter-kit/internal/pkg/observability/logger"
)

const (
	httpServerShutdownTimeout = 15 * time.Second
	HTTPWorkerName            = "http-worker"
)

// HTTPWorker runs an http.Server with graceful shutdown on SIGINT/SIGTERM.
type HTTPWorker struct {
	srv *http.Server
	lgr logger.Logger
}

func NewHTTPWorker(srv *http.Server, lgr logger.Logger) (*HTTPWorker, error) {
	if srv == nil {
		return nil, fmt.Errorf("server is required to create %s", HTTPWorkerName)
	}
	if lgr == nil {
		return nil, fmt.Errorf("logger is required to create %s", HTTPWorkerName)
	}
	return &HTTPWorker{srv: srv, lgr: lgr}, nil
}

func (h *HTTPWorker) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(stop)

	errCh := make(chan error, 1)

	h.lgr.Info(ctx, fmt.Sprintf("HTTP server listening on %s", h.srv.Addr))
	go func() {
		if err := h.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case sig := <-stop:
		h.lgr.Info(ctx, fmt.Sprintf("shutdown signal received: %s", sig))
	case err := <-errCh:
		if err != nil {
			h.lgr.Error(ctx, err, "HTTP server stopped with error")
			return err
		}
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), httpServerShutdownTimeout)
	defer shutdownCancel()

	h.lgr.Info(ctx, "shutting down HTTP server...")
	if err := h.srv.Shutdown(shutdownCtx); err != nil {
		h.lgr.Error(ctx, err, "failed to shutdown HTTP server gracefully")
		return err
	}
	h.lgr.Info(ctx, "HTTP server shutdown complete")
	return nil
}
