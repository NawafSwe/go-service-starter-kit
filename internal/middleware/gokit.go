package middleware

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/go-kit/kit/endpoint"
)

// DeadlineError is returned when a context deadline is exceeded inside an endpoint.
type DeadlineError struct{}

func (DeadlineError) Error() string { return "context deadline exceeded" }

var ErrDeadlineError = DeadlineError{}

func NewDeadlineError() DeadlineError { return DeadlineError{} }

// TimeoutMiddleware wraps an endpoint with a context deadline.
// A zero or negative timeout is treated as disabled.
func TimeoutMiddleware(timeout time.Duration) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request any) (any, error) {
			if timeout <= 0 {
				return next(ctx, request)
			}
			ctx, cancel := context.WithDeadline(ctx, time.Now().Add(timeout))
			defer cancel()

			response, err := next(ctx, request)
			if errors.Is(err, context.DeadlineExceeded) {
				return nil, NewDeadlineError()
			}
			return response, err
		}
	}
}

// ErrRateLimitReached is returned when a rate limit is exceeded.
var ErrRateLimitReached = rateLimitError{}

type rateLimitError struct{}

func (rateLimitError) Error() string { return "too many requests" }

// RateLimit is a sliding-window rate limiter middleware.
// A zero limit or interval disables rate limiting.
func RateLimit(interval time.Duration, limit int64) endpoint.Middleware {
	var roundCounter int64
	lastCheckpoint := time.Now().UnixMilli()

	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request any) (any, error) {
			if limit <= 0 || interval <= 0 {
				return next(ctx, request)
			}
			if time.Now().UnixMilli()-atomic.LoadInt64(&lastCheckpoint) >= interval.Milliseconds() {
				atomic.StoreInt64(&roundCounter, 0)
				atomic.StoreInt64(&lastCheckpoint, time.Now().UnixMilli())
			}
			atomic.AddInt64(&roundCounter, 1)
			if atomic.LoadInt64(&roundCounter) > limit {
				return nil, ErrRateLimitReached
			}
			return next(ctx, request)
		}
	}
}
