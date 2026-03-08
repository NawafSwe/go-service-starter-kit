package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/observability/logger"
)

type (
	// EncodeDecoder combines Encoder and Decoder.
	EncodeDecoder interface {
		Encoder
		Decoder
	}

	// Decoder decodes an HTTP request into a domain request type.
	Decoder interface {
		Decode(ctx context.Context, httpRequest *http.Request) (any, error)
	}

	// Encoder encodes a domain response type into an HTTP response.
	Encoder interface {
		Encode(ctx context.Context, responseWriter http.ResponseWriter, response any) error
	}

	// ErrorEncoder encodes an error into an HTTP response.
	ErrorEncoder interface {
		Encode(ctx context.Context, err error, responseWriter http.ResponseWriter)
	}
)

// MakeHTTPHandler assembles a go-kit HTTP server with the provided middlewares applied in order.
func MakeHTTPHandler(
	ep endpoint.Endpoint,
	decoder Decoder,
	encoder Encoder,
	errorEncoder ErrorEncoder,
	lgr logger.Logger,
	middlewares ...endpoint.Middleware,
) http.Handler {
	for _, mw := range middlewares {
		ep = mw(ep)
	}
	return kithttp.NewServer(
		ep,
		decoder.Decode,
		encoder.Encode,
		kithttp.ServerErrorEncoder(errorEncoder.Encode),
		kithttp.ServerErrorHandler(NewLogErrorHandler(lgr)),
	)
}

// LogErrorHandler logs errors produced inside go-kit HTTP servers.
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
	h.lgr.Error(ctx, err, "error in HTTP handler")
}
