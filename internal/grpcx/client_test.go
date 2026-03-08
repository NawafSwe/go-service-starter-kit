package grpcx_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nawafswe/go-service-starter-kit/internal/grpcx"
	"github.com/nawafswe/go-service-starter-kit/internal/grpcx/mock"
	"github.com/sony/gobreaker"
	"go.opentelemetry.io/otel/metric/noop"
	nooptrace "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/mock/gomock"
)

func defaultConfig() grpcx.Config {
	return grpcx.Config{
		Name:         "test-grpc-service",
		Address:      "localhost:50051",
		Timeout:      5 * time.Second,
		MaxRetries:   2,
		RetryWaitMin: 0,
		RetryWaitMax: 0,
		CircuitBreaker: grpcx.CircuitBreakerConfig{
			MaxRequests: 5,
			Interval:    0,
			Timeout:     time.Second,
			Threshold:   5,
		},
	}
}

func newClient(t *testing.T, cfg grpcx.Config, invoker grpcx.Invoker) *grpcx.Client {
	t.Helper()
	c, err := grpcx.New(cfg, invoker, noop.NewMeterProvider().Meter("test"), nooptrace.NewTracerProvider())
	if err != nil {
		t.Fatalf("grpcx.New: %v", err)
	}
	return c
}

func TestNew(t *testing.T) {
	ctrl := gomock.NewController(t)
	c, err := grpcx.New(defaultConfig(), mock.NewMockInvoker(ctrl), noop.NewMeterProvider().Meter("test"), nooptrace.NewTracerProvider())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestClient_Invoke(t *testing.T) {
	const method = "/example.v1.ExampleService/GetExample"

	tests := []struct {
		name    string
		setup   func(m *mock.MockInvoker)
		cfg     func() grpcx.Config
		wantErr bool
	}{
		{
			name: "success on first attempt",
			setup: func(m *mock.MockInvoker) {
				m.EXPECT().Invoke(gomock.Any(), method, gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			name: "retries on error then succeeds",
			setup: func(m *mock.MockInvoker) {
				gomock.InOrder(
					m.EXPECT().Invoke(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("transient")),
					m.EXPECT().Invoke(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil),
				)
			},
			cfg: func() grpcx.Config {
				c := defaultConfig()
				c.MaxRetries = 1
				return c
			},
		},
		{
			name: "exhausts all retries",
			setup: func(m *mock.MockInvoker) {
				m.EXPECT().Invoke(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("persistent")).Times(3)
			},
			wantErr: true,
		},
		{
			name: "context cancelled during retry",
			setup: func(m *mock.MockInvoker) {
				m.EXPECT().Invoke(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("err")).AnyTimes()
			},
			cfg: func() grpcx.Config {
				c := defaultConfig()
				c.MaxRetries = 3
				c.RetryWaitMin = 50 * time.Millisecond
				c.RetryWaitMax = 100 * time.Millisecond
				return c
			},
			wantErr: true,
		},
		{
			name: "threshold zero never opens circuit breaker",
			setup: func(m *mock.MockInvoker) {
				m.EXPECT().Invoke(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("err")).Times(1)
			},
			cfg: func() grpcx.Config {
				c := defaultConfig()
				c.MaxRetries = 0
				c.CircuitBreaker = grpcx.CircuitBreakerConfig{Threshold: 0, Timeout: time.Second}
				return c
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			m := mock.NewMockInvoker(ctrl)
			tc.setup(m)

			cfg := defaultConfig()
			if tc.cfg != nil {
				cfg = tc.cfg()
			}
			c := newClient(t, cfg, m)

			ctx := context.Background()
			if tc.name == "context cancelled during retry" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			err := c.Invoke(ctx, method, nil, nil)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestClient_Invoke_CircuitBreakerOpen(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := mock.NewMockInvoker(ctrl)
	m.EXPECT().Invoke(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(errors.New("err")).Times(1)

	cfg := defaultConfig()
	cfg.MaxRetries = 0
	cfg.CircuitBreaker = grpcx.CircuitBreakerConfig{MaxRequests: 1, Timeout: time.Second, Threshold: 1}
	c := newClient(t, cfg, m)

	c.Invoke(context.Background(), "/svc/Method", nil, nil)

	err := c.Invoke(context.Background(), "/svc/Method", nil, nil)
	if !errors.Is(err, gobreaker.ErrOpenState) {
		t.Errorf("expected ErrOpenState, got %v", err)
	}
}
