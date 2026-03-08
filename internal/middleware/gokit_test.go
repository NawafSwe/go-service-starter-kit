package middleware_test

import (
	"context"
	"testing"
	"time"

	"github.com/nawafswe/go-service-starter-kit/internal/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimeoutMiddleware(t *testing.T) {
	tests := []struct {
		name        string
		timeout     time.Duration
		ep          func(context.Context, any) (any, error)
		expectedErr error
		wantResp    any
	}{
		{
			name:     "disabled when timeout is zero",
			timeout:  0,
			ep:       func(_ context.Context, _ any) (any, error) { return "ok", nil },
			wantResp: "ok",
		},
		{
			name:    "exceeds deadline",
			timeout: 1 * time.Millisecond,
			ep: func(ctx context.Context, _ any) (any, error) {
				<-ctx.Done()
				return nil, ctx.Err()
			},
			expectedErr: middleware.ErrDeadlineError,
		},
		{
			name:     "within deadline",
			timeout:  time.Second,
			ep:       func(_ context.Context, _ any) (any, error) { return "done", nil },
			wantResp: "done",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wrapped := middleware.TimeoutMiddleware(tc.timeout)(tc.ep)
			resp, err := wrapped(context.Background(), nil)
			if tc.expectedErr != nil {
				assert.ErrorIs(t, err, tc.expectedErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantResp, resp)
		})
	}
}

func TestRateLimit_Disabled(t *testing.T) {
	ep := func(_ context.Context, req any) (any, error) { return "ok", nil }

	tests := []struct {
		name     string
		interval time.Duration
		limit    int64
	}{
		{name: "zero limit", interval: time.Second, limit: 0},
		{name: "zero interval", interval: 0, limit: 10},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wrapped := middleware.RateLimit(tc.interval, tc.limit)(ep)
			for range 5 {
				_, err := wrapped(context.Background(), nil)
				require.NoError(t, err)
			}
		})
	}
}

func TestRateLimit_ExceedsLimit(t *testing.T) {
	ep := func(_ context.Context, req any) (any, error) { return "ok", nil }
	wrapped := middleware.RateLimit(time.Hour, 2)(ep)

	// First two calls should succeed.
	for range 2 {
		_, err := wrapped(context.Background(), nil)
		require.NoError(t, err)
	}
	// Third call should be rate limited.
	_, err := wrapped(context.Background(), nil)
	assert.ErrorIs(t, err, middleware.ErrRateLimitReached)
}
