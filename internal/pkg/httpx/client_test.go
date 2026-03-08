package httpx_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/nawafswe/go-service-starter-kit/internal/pkg/httpx"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/httpx/mock"
	"github.com/sony/gobreaker"
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

func newClient(t *testing.T, cfg httpx.Config, doer httpx.Doer) *httpx.Client {
	t.Helper()
	c, err := httpx.New(cfg, doer, noop.NewMeterProvider().Meter("test"), nooptrace.NewTracerProvider())
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
	ctrl := gomock.NewController(t)
	c, err := httpx.New(defaultConfig(), mock.NewMockDoer(ctrl), noop.NewMeterProvider().Meter("test"), nooptrace.NewTracerProvider())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestClient_Do(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(m *mock.MockDoer)
		cfg     func() httpx.Config
		wantErr bool
	}{
		{
			name: "success on first attempt",
			setup: func(m *mock.MockDoer) {
				m.EXPECT().Do(gomock.Any()).Return(okResponse(), nil)
			},
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
			c := newClient(t, cfg, m)

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
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.StatusCode != http.StatusOK {
				t.Errorf("expected 200, got %d", resp.StatusCode)
			}
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

type errReader struct{}

func (e *errReader) Read(_ []byte) (int, error) { return 0, errors.New("read error") }
func (e *errReader) Close() error               { return nil }
