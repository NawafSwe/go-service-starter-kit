package grpcx

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/sony/gobreaker"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

//go:generate mockgen -destination=mock/mock.go -package=mock github.com/nawafswe/go-service-starter-kit/internal/grpcx Invoker

type Invoker interface {
	Invoke(ctx context.Context, method string, args, reply any, opts ...grpc.CallOption) error
}

// Option configures the resilient gRPC client.
type Option func(*Client)

// WithMeter enables OTel metrics on the gRPC client.
func WithMeter(m metric.Meter) Option {
	return func(c *Client) { c.meter = m }
}

// WithTracerProvider enables OTel tracing on the gRPC client.
func WithTracerProvider(tp trace.TracerProvider) Option {
	return func(c *Client) { c.tracer = tp.Tracer(c.cfg.Name) }
}

type Client struct {
	invoker Invoker
	cb      *gobreaker.CircuitBreaker
	cfg     Config
	meter   metric.Meter
	counter metric.Int64Counter
	tracer  trace.Tracer
}

// New creates a resilient gRPC client with circuit breaker and retry support.
// By default, tracing and metrics are disabled. Use WithMeter and
// WithTracerProvider to enable them.
func New(cfg Config, invoker Invoker, opts ...Option) (*Client, error) {
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
		invoker: invoker,
		cb:      cb,
		cfg:     cfg,
	}
	for _, opt := range opts {
		opt(c)
	}

	if c.meter != nil {
		counter, err := c.meter.Int64Counter(
			"grpc_client_requests_total",
			metric.WithDescription("Total gRPC client requests by dependency"),
		)
		if err != nil {
			return nil, fmt.Errorf("grpcx: register counter: %w", err)
		}
		c.counter = counter
	}

	return c, nil
}

func (c *Client) Invoke(ctx context.Context, method string, args, reply any, opts ...grpc.CallOption) error {
	var span trace.Span
	if c.tracer != nil {
		ctx, span = c.tracer.Start(ctx, method, trace.WithSpanKind(trace.SpanKindClient))
		defer span.End()
	}

	baseAttrs := []attribute.KeyValue{
		attribute.String("dependency", c.cfg.Name),
		attribute.String("circuit_breaker", c.cb.Name()),
		attribute.String("method", method),
	}

	var lastErr error
	for attempt := 0; attempt <= c.cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(jitterBackoff(c.cfg.RetryWaitMin, c.cfg.RetryWaitMax, attempt)):
			}
		}

		_, err := c.cb.Execute(func() (any, error) {
			return nil, c.invoker.Invoke(ctx, method, args, reply, opts...)
		})

		if err == nil {
			if c.counter != nil {
				c.counter.Add(ctx, 1, metric.WithAttributes(
					append(baseAttrs, attribute.String("result", "success"))...,
				))
			}
			return nil
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
	return lastErr
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
