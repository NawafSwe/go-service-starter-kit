package grpc

import (
	"context"
	"errors"

	"github.com/go-kit/kit/endpoint"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/observability/logger"
)

// MakeGRPCHandler assembles a go-kit gRPC server handler with the provided middlewares applied in order.
func MakeGRPCHandler(
	ep endpoint.Endpoint,
	decoder grpctransport.DecodeRequestFunc,
	encoder grpctransport.EncodeResponseFunc,
	lgr logger.Logger,
	middlewares ...endpoint.Middleware,
) grpctransport.Handler {
	for _, mw := range middlewares {
		ep = mw(ep)
	}
	return grpctransport.NewServer(
		ep,
		decoder,
		encoder,
		grpctransport.ServerErrorHandler(NewLogErrorHandler(lgr)),
	)
}

// LogErrorHandler logs errors produced inside go-kit gRPC handlers.
type LogErrorHandler struct {
	lgr logger.Logger
}

func NewLogErrorHandler(lgr logger.Logger) *LogErrorHandler {
	return &LogErrorHandler{lgr: lgr}
}

func (h LogErrorHandler) Handle(ctx context.Context, err error) {
	if errors.Is(ctx.Err(), context.Canceled) || errors.Is(err, context.Canceled) {
		return
	}
	h.lgr.Error(ctx, err, "error in gRPC handler")
}
