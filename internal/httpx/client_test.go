package httpx_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/nawafswe/go-service-starter-kit/internal/httpx"
	"github.com/nawafswe/go-service-starter-kit/internal/httpx/mock"
	"github.com/sony/gobreaker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric/noop"
	nooptrace "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/mock/gomock"
)

func defaultConfig() httpx.Config {
	return httpx.Config{
		Name:         "test-service",
		BaseURL:      "http://example.com",
		Timeout:      5 * time.Second,
		MaxRetries:   2,
		RetryWaitMin: 0,
		RetryWaitMax: 0,
		CircuitBreaker: httpx.CircuitBreakerConfig{
			MaxRequests: 5,
			Interval:    0,
			Timeout:     time.Second,
			Threshold:   5,
		},
	}
}

func newClient(t *testing.T, cfg httpx.Config, doer httpx.Doer, opts ...httpx.Option) *httpx.Client {
	t.Helper()
	c, err := httpx.New(cfg, doer, opts...)
	if err != nil {
		t.Fatalf("httpx.New: %v", err)
	}
	return c
}

func okResponse() *http.Response {
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBufferString(`{}`))}
}

func serverErrResponse() *http.Response {
	return &http.Response{StatusCode: http.StatusInternalServerError, Body: io.NopCloser(bytes.NewBufferString("err"))}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		opts []httpx.Option
	}{
		{
			name: "without options",
		},
		{
			name: "with meter",
			opts: []httpx.Option{httpx.WithMeter(noop.NewMeterProvider().Meter("test"))},
		},
		{
			name: "with tracer provider",
			opts: []httpx.Option{httpx.WithTracerProvider(nooptrace.NewTracerProvider())},
		},
		{
			name: "with meter and tracer provider",
			opts: []httpx.Option{
				httpx.WithMeter(noop.NewMeterProvider().Meter("test")),
				httpx.WithTracerProvider(nooptrace.NewTracerProvider()),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			c, err := httpx.New(defaultConfig(), mock.NewMockDoer(ctrl), tc.opts...)
			require.NoError(t, err)
			assert.NotNil(t, c)
		})
	}
}

func TestClient_Do(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(m *mock.MockDoer)
		cfg     func() httpx.Config
		opts    []httpx.Option
		wantErr bool
	}{
		{
			name: "success on first attempt",
			setup: func(m *mock.MockDoer) {
				m.EXPECT().Do(gomock.Any()).Return(okResponse(), nil)
			},
		},
		{
			name: "success with meter and tracer",
			setup: func(m *mock.MockDoer) {
				m.EXPECT().Do(gomock.Any()).Return(okResponse(), nil)
			},
			opts: []httpx.Option{
				httpx.WithMeter(noop.NewMeterProvider().Meter("test")),
				httpx.WithTracerProvider(nooptrace.NewTracerProvider()),
			},
		},
		{
			name: "error with meter and tracer records metrics",
			setup: func(m *mock.MockDoer) {
				m.EXPECT().Do(gomock.Any()).Return(nil, errors.New("err")).Times(1)
			},
			cfg: func() httpx.Config {
				c := defaultConfig()
				c.MaxRetries = 0
				return c
			},
			opts: []httpx.Option{
				httpx.WithMeter(noop.NewMeterProvider().Meter("test")),
				httpx.WithTracerProvider(nooptrace.NewTracerProvider()),
			},
			wantErr: true,
		},
		{
			name: "retries on server error then succeeds",
			setup: func(m *mock.MockDoer) {
				gomock.InOrder(
					m.EXPECT().Do(gomock.Any()).Return(serverErrResponse(), nil),
					m.EXPECT().Do(gomock.Any()).Return(okResponse(), nil),
				)
			},
			cfg: func() httpx.Config {
				c := defaultConfig()
				c.MaxRetries = 1
				return c
			},
		},
		{
			name: "exhausts all retries",
			setup: func(m *mock.MockDoer) {
				m.EXPECT().Do(gomock.Any()).Return(serverErrResponse(), nil).Times(3)
			},
			wantErr: true,
		},
		{
			name: "network error",
			setup: func(m *mock.MockDoer) {
				m.EXPECT().Do(gomock.Any()).Return(nil, errors.New("connection refused"))
			},
			cfg: func() httpx.Config {
				c := defaultConfig()
				c.MaxRetries = 0
				return c
			},
			wantErr: true,
		},
		{
			name: "context cancelled during retry",
			setup: func(m *mock.MockDoer) {
				m.EXPECT().Do(gomock.Any()).Return(serverErrResponse(), nil).AnyTimes()
			},
			cfg: func() httpx.Config {
				c := defaultConfig()
				c.MaxRetries = 3
				c.RetryWaitMin = 50 * time.Millisecond
				c.RetryWaitMax = 100 * time.Millisecond
				return c
			},
			wantErr: true,
		},
		{
			name: "body reset on retry",
			setup: func(m *mock.MockDoer) {
				gomock.InOrder(
					m.EXPECT().Do(gomock.Any()).Return(serverErrResponse(), nil),
					m.EXPECT().Do(gomock.Any()).DoAndReturn(func(req *http.Request) (*http.Response, error) {
						body, _ := io.ReadAll(req.Body)
						if string(body) != `{"name":"test"}` {
							return nil, errors.New("body not reset")
						}
						return okResponse(), nil
					}),
				)
			},
			cfg: func() httpx.Config {
				c := defaultConfig()
				c.MaxRetries = 1
				return c
			},
		},
		{
			name: "threshold zero never opens circuit breaker",
			setup: func(m *mock.MockDoer) {
				m.EXPECT().Do(gomock.Any()).Return(nil, errors.New("err")).Times(1)
			},
			cfg: func() httpx.Config {
				c := defaultConfig()
				c.MaxRetries = 0
				c.CircuitBreaker = httpx.CircuitBreakerConfig{Threshold: 0, Timeout: time.Second}
				return c
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			m := mock.NewMockDoer(ctrl)
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

			body := io.NopCloser(bytes.NewBufferString(`{"name":"test"}`))
			var req *http.Request
			if tc.name == "body reset on retry" {
				req, _ = http.NewRequest(http.MethodPost, "http://example.com/items", body)
			} else {
				req, _ = http.NewRequest(http.MethodGet, "http://example.com/items", nil)
			}

			resp, err := c.Do(ctx, req)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestClient_Do_CircuitBreakerOpen(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := mock.NewMockDoer(ctrl)
	m.EXPECT().Do(gomock.Any()).Return(nil, errors.New("err")).Times(1)

	cfg := defaultConfig()
	cfg.MaxRetries = 0
	cfg.CircuitBreaker = httpx.CircuitBreakerConfig{MaxRequests: 1, Timeout: time.Second, Threshold: 1}
	c := newClient(t, cfg, m)

	req1, _ := http.NewRequest(http.MethodGet, "http://example.com/items", nil)
	c.Do(context.Background(), req1)

	req2, _ := http.NewRequest(http.MethodGet, "http://example.com/items", nil)
	_, err := c.Do(context.Background(), req2)
	if !errors.Is(err, gobreaker.ErrOpenState) {
		t.Errorf("expected ErrOpenState, got %v", err)
	}
}

func TestClient_Do_BodyReadError(t *testing.T) {
	ctrl := gomock.NewController(t)
	c := newClient(t, defaultConfig(), mock.NewMockDoer(ctrl))
	req, _ := http.NewRequest(http.MethodPost, "http://example.com/items", &errReader{})
	_, err := c.Do(context.Background(), req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_Do_RetryBackoff(t *testing.T) {
	tests := []struct {
		name string
		cfg  func() httpx.Config
	}{
		{
			name: "jitter backoff with min greater than zero",
			cfg: func() httpx.Config {
				c := defaultConfig()
				c.MaxRetries = 1
				c.RetryWaitMin = 1 * time.Millisecond
				c.RetryWaitMax = 10 * time.Millisecond
				return c
			},
		},
		{
			name: "backoff capped at max when base exceeds max",
			cfg: func() httpx.Config {
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
			m := mock.NewMockDoer(ctrl)
			gomock.InOrder(
				m.EXPECT().Do(gomock.Any()).Return(serverErrResponse(), nil),
				m.EXPECT().Do(gomock.Any()).Return(okResponse(), nil),
			)
			c := newClient(t, tc.cfg(), m)
			req, _ := http.NewRequest(http.MethodGet, "http://example.com/items", nil)
			resp, err := c.Do(context.Background(), req)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

type errReader struct{}

func (e *errReader) Read(_ []byte) (int, error) { return 0, errors.New("read error") }
func (e *errReader) Close() error               { return nil }
