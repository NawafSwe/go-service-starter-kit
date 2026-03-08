package http_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-kit/kit/endpoint"
	gohttp "github.com/nawafswe/go-service-starter-kit/internal/gokit/http"
	"github.com/nawafswe/go-service-starter-kit/internal/observability/logger"
	"github.com/stretchr/testify/assert"
)

// stubDecoder decodes nothing and returns nil.
type stubDecoder struct{}

func (stubDecoder) Decode(_ context.Context, _ *http.Request) (any, error) { return nil, nil }

// stubEncoder writes a 200 with no body.
type stubEncoder struct{}

func (stubEncoder) Encode(_ context.Context, w http.ResponseWriter, _ any) error {
	w.WriteHeader(http.StatusOK)
	return nil
}

// stubErrorEncoder writes a 500.
type stubErrorEncoder struct{}

func (stubErrorEncoder) Encode(_ context.Context, _ error, w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
}

func newTestLogger() logger.ZerologLogger {
	return logger.NewLogger(logger.ErrorLevel, "test", "0.0.0", "test")
}

func TestMakeHTTPHandler(t *testing.T) {
	tests := []struct {
		name       string
		mw         []endpoint.Middleware
		wantCalled bool
	}{
		{name: "success without middleware"},
		{
			name: "success with middleware",
			mw: []endpoint.Middleware{
				func(next endpoint.Endpoint) endpoint.Endpoint {
					return func(ctx context.Context, req any) (any, error) {
						return next(ctx, req)
					}
				},
			},
			wantCalled: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			var mws []endpoint.Middleware
			if tc.wantCalled {
				mws = []endpoint.Middleware{
					func(next endpoint.Endpoint) endpoint.Endpoint {
						return func(ctx context.Context, req any) (any, error) {
							called = true
							return next(ctx, req)
						}
					},
				}
			}

			ep := func(_ context.Context, _ any) (any, error) { return "ok", nil }
			handler := gohttp.MakeHTTPHandler(ep, stubDecoder{}, stubEncoder{}, stubErrorEncoder{}, newTestLogger(), mws...)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)
			if tc.wantCalled {
				assert.True(t, called, "middleware was not called")
			}
		})
	}
}

func TestLogErrorHandler(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
		err  error
	}{
		{name: "regular error", ctx: context.Background(), err: errors.New("some error")},
		{name: "context cancelled", ctx: func() context.Context {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			return ctx
		}(), err: context.Canceled},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := gohttp.NewLogErrorHandler(newTestLogger())
			assert.NotPanics(t, func() { h.Handle(tc.ctx, tc.err) })
		})
	}
}
