package middleware_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/go-kit/kit/endpoint"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/middleware"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/observability/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func discardLogger() logger.ZerologLogger {
	return logger.NewLogger(logger.DebugLevel, "test", "0.0.0", "test").WithOutput(&bytes.Buffer{})
}

// stubRequest is a struct used as a request/response in logging tests.
// go-masker requires struct values — strings and primitives will panic.
type stubRequest struct {
	Name string
}

type stubResponse struct {
	OK bool
}

func TestLoggingMiddleware(t *testing.T) {
	wantErr := errors.New("something failed")

	tests := []struct {
		name        string
		mw          endpoint.Middleware
		ep          endpoint.Endpoint
		expectedErr error
	}{
		{
			name: "HTTP success",
			mw:   middleware.LoggingHTTPMiddleware(discardLogger(), "/test", "GET"),
			ep:   func(_ context.Context, _ any) (any, error) { return stubResponse{OK: true}, nil },
		},
		{
			name:        "HTTP error",
			mw:          middleware.LoggingHTTPMiddleware(discardLogger(), "/test", "POST"),
			ep:          func(_ context.Context, _ any) (any, error) { return nil, wantErr },
			expectedErr: wantErr,
		},
		{
			name: "gRPC success",
			mw:   middleware.LoggingGRPCMiddleware(discardLogger(), "SomeRPC"),
			ep:   func(_ context.Context, _ any) (any, error) { return stubResponse{OK: true}, nil },
		},
		{
			name: "PubSub success",
			mw:   middleware.LoggingPubSubMiddleware(discardLogger(), "example.created"),
			ep:   func(_ context.Context, _ any) (any, error) { return stubResponse{OK: true}, nil },
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wrapped := tc.mw(tc.ep)
			resp, err := wrapped(context.Background(), stubRequest{Name: "test"})
			if tc.expectedErr != nil {
				assert.ErrorIs(t, err, tc.expectedErr)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, resp)
		})
	}
}
