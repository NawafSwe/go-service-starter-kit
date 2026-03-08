package httpx

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/sony/gobreaker"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

//go:generate mockgen -destination=mock/mock.go -package=mock github.com/nawafswe/go-service-starter-kit/internal/httpx Doer

type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Option configures the resilient HTTP client.
type Option func(*Client)

// WithMeter enables OTel metrics on the HTTP client.
func WithMeter(m metric.Meter) Option {
	return func(c *Client) { c.meter = m }
}

// WithTracerProvider enables OTel tracing on the HTTP client.
func WithTracerProvider(tp trace.TracerProvider) Option {
	return func(c *Client) { c.tracer = tp.Tracer(c.cfg.Name) }
}

type Client struct {
	doer    Doer
	cb      *gobreaker.CircuitBreaker
	cfg     Config
	meter   metric.Meter
	counter metric.Int64Counter
	tracer  trace.Tracer
}

// New creates a resilient HTTP client with circuit breaker and retry support.
// By default, tracing and metrics are disabled. Use WithMeter and
// WithTracerProvider to enable them.
func New(cfg Config, doer Doer, opts ...Option) (*Client, error) {
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        cfg.Name,
		MaxRequests: cfg.CircuitBreaker.MaxRequests,
		Interval:    cfg.CircuitBreaker.Interval,
		Timeout:     cfg.CircuitBreaker.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return cfg.CircuitBreaker.Threshold > 0 &&
				counts.ConsecutiveFailures >= cfg.CircuitBreaker.Threshold
		},
	})

	c := &Client{
		doer: doer,
		cb:   cb,
		cfg:  cfg,
	}
	for _, opt := range opts {
		opt(c)
	}

	if c.meter != nil {
		counter, err := c.meter.Int64Counter(
			"http_client_requests_total",
			metric.WithDescription("Total HTTP client requests by dependency"),
		)
		if err != nil {
			return nil, fmt.Errorf("httpx: register counter: %w", err)
		}
		c.counter = counter
	}

	return c, nil
}

func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	var bodyBytes []byte
	if req.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("httpx: read body: %w", err)
		}
		req.Body.Close()
	}

	var span trace.Span
	if c.tracer != nil {
		ctx, span = c.tracer.Start(ctx, fmt.Sprintf("%s %s", req.Method, req.URL.Path),
			trace.WithSpanKind(trace.SpanKindClient),
		)
		defer span.End()
	}

	baseAttrs := []attribute.KeyValue{
		attribute.String("dependency", c.cfg.Name),
		attribute.String("circuit_breaker", c.cb.Name()),
		attribute.String("method", req.Method),
	}

	var lastErr error
	for attempt := 0; attempt <= c.cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(jitterBackoff(c.cfg.RetryWaitMin, c.cfg.RetryWaitMax, attempt)):
			}
		}

		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			req.ContentLength = int64(len(bodyBytes))
		}

		result, err := c.cb.Execute(func() (any, error) {
			resp, e := c.doer.Do(req.WithContext(ctx))
			if e != nil {
				return nil, e
			}
			if resp.StatusCode >= 500 {
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
				return nil, fmt.Errorf("server error: status %d", resp.StatusCode)
			}
			return resp, nil
		})

		if err == nil {
			if c.counter != nil {
				c.counter.Add(ctx, 1, metric.WithAttributes(
					append(baseAttrs, attribute.String("result", "success"))...,
				))
			}
			return result.(*http.Response), nil
		}

		lastErr = err
		if c.counter != nil {
			c.counter.Add(ctx, 1, metric.WithAttributes(
				append(baseAttrs, attribute.String("result", "error"))...,
			))
		}

		if errors.Is(err, gobreaker.ErrOpenState) || errors.Is(err, gobreaker.ErrTooManyRequests) {
			break
		}
	}

	if span != nil {
		span.RecordError(lastErr)
	}
	return nil, lastErr
}

func jitterBackoff(min, max time.Duration, attempt int) time.Duration {
	if min <= 0 {
		return 0
	}
	base := min * (1 << uint(attempt-1))
	if base > max || base <= 0 {
		base = max
	}
	if base <= 0 {
		return 0
	}
	return base + time.Duration(rand.Int63n(int64(base)))
}
