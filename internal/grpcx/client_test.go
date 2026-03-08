package grpcx_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nawafswe/go-service-starter-kit/internal/grpcx"
	"github.com/nawafswe/go-service-starter-kit/internal/grpcx/mock"
	"github.com/sony/gobreaker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func newClient(t *testing.T, cfg grpcx.Config, invoker grpcx.Invoker, opts ...grpcx.Option) *grpcx.Client {
	t.Helper()
	c, err := grpcx.New(cfg, invoker, opts...)
	if err != nil {
		t.Fatalf("grpcx.New: %v", err)
	}
	return c
}

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		opts []grpcx.Option
	}{
		{
			name: "without options",
		},
		{
			name: "with meter",
			opts: []grpcx.Option{grpcx.WithMeter(noop.NewMeterProvider().Meter("test"))},
		},
		{
			name: "with tracer provider",
			opts: []grpcx.Option{grpcx.WithTracerProvider(nooptrace.NewTracerProvider())},
		},
		{
			name: "with meter and tracer provider",
			opts: []grpcx.Option{
				grpcx.WithMeter(noop.NewMeterProvider().Meter("test")),
				grpcx.WithTracerProvider(nooptrace.NewTracerProvider()),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			c, err := grpcx.New(defaultConfig(), mock.NewMockInvoker(ctrl), tc.opts...)
			require.NoError(t, err)
			assert.NotNil(t, c)
		})
	}
}

func TestClient_Invoke(t *testing.T) {
	const method = "/example.v1.ExampleService/GetExample"

	tests := []struct {
		name    string
		setup   func(m *mock.MockInvoker)
		cfg     func() grpcx.Config
		opts    []grpcx.Option
		wantErr bool
	}{
		{
			name: "success on first attempt",
			setup: func(m *mock.MockInvoker) {
				m.EXPECT().Invoke(gomock.Any(), method, gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			name: "success with meter and tracer",
			setup: func(m *mock.MockInvoker) {
				m.EXPECT().Invoke(gomock.Any(), method, gomock.Any(), gomock.Any()).Return(nil)
			},
			opts: []grpcx.Option{
				grpcx.WithMeter(noop.NewMeterProvider().Meter("test")),
				grpcx.WithTracerProvider(nooptrace.NewTracerProvider()),
			},
		},
		{
			name: "error with meter and tracer records metrics",
			setup: func(m *mock.MockInvoker) {
				m.EXPECT().Invoke(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("err")).Times(1)
			},
			cfg: func() grpcx.Config {
				c := defaultConfig()
				c.MaxRetries = 0
				return c
			},
			opts: []grpcx.Option{
				grpcx.WithMeter(noop.NewMeterProvider().Meter("test")),
				grpcx.WithTracerProvider(nooptrace.NewTracerProvider()),
			},
			wantErr: true,
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
			c := newClient(t, cfg, m, tc.opts...)

			ctx := context.Background()
			if tc.name == "context cancelled during retry" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			err := c.Invoke(ctx, method, nil, nil)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestClient_Invoke_RetryBackoff(t *testing.T) {
	const method = "/example.v1.ExampleService/GetExample"

	tests := []struct {
		name string
		cfg  func() grpcx.Config
	}{
		{
			name: "jitter backoff with min greater than zero",
			cfg: func() grpcx.Config {
				c := defaultConfig()
				c.MaxRetries = 1
				c.RetryWaitMin = 1 * time.Millisecond
				c.RetryWaitMax = 10 * time.Millisecond
				return c
			},
		},
		{
			name: "backoff capped at max when base exceeds max",
			cfg: func() grpcx.Config {
				c := defaultConfig()
				c.MaxRetries = 1
				c.RetryWaitMin = 100 * time.Millisecond
				c.RetryWaitMax = 50 * time.Millisecond
				return c
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			m := mock.NewMockInvoker(ctrl)
			gomock.InOrder(
				m.EXPECT().Invoke(gomock.Any(), method, gomock.Any(), gomock.Any()).Return(errors.New("transient")),
				m.EXPECT().Invoke(gomock.Any(), method, gomock.Any(), gomock.Any()).Return(nil),
			)
			c := newClient(t, tc.cfg(), m)
			err := c.Invoke(context.Background(), method, nil, nil)
			require.NoError(t, err)
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
