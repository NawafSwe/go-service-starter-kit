package metric_test

import (
	"context"
	"testing"

	"github.com/nawafswe/go-service-starter-kit/internal/observability/metric"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric/noop"
)

func TestReporter(t *testing.T) {
	tests := []struct {
		name  string
		attrs map[string]any
	}{
		{
			name: "record with known attr types",
			attrs: map[string]any{
				"endpoint": "/health",
				"method":   "GET",
				"status":   200,
				"success":  true,
			},
		},
		{
			name: "record with int64 attr type",
			attrs: map[string]any{
				"latency_ns": int64(123456),
			},
		},
		{
			name:  "record with nil map",
			attrs: nil,
		},
		{
			name: "record with unknown attr types",
			attrs: map[string]any{
				"known":   "value",
				"unknown": []byte("ignored"),
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := noop.NewMeterProvider().Meter("test")
			r, err := metric.NewReporter(m)
			require.NoError(t, err)
			assert.NotNil(t, r)

			assert.NotPanics(t, func() {
				r.RecordRequest(context.Background(), tc.attrs)
			})
		})
	}
}
