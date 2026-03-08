package tracing_test

import (
	"context"
	"errors"
	"testing"

	"github.com/nawafswe/go-service-starter-kit/internal/pkg/config"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/observability/tracing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	nooptrace "go.opentelemetry.io/otel/trace/noop"
)

func TestSetup_Disabled(t *testing.T) {
	cfg := config.Config{}
	cfg.General.Tracing.Enabled = false

	tp, shutdown, err := tracing.Setup(context.Background(), cfg)
	require.NoError(t, err)
	assert.NotNil(t, tp)
	assert.NoError(t, shutdown(context.Background()))
}

func TestStartSpan(t *testing.T) {
	tests := []struct {
		name  string
		attrs []attribute.KeyValue
	}{
		{name: "without attributes", attrs: nil},
		{name: "with attributes", attrs: []attribute.KeyValue{
			attribute.String("service", "test"),
			attribute.Int64("count", 1),
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx, span := tracing.StartSpan(context.Background(), "test-tracer", "test-op", tc.attrs)
			assert.NotNil(t, ctx)
			assert.NotNil(t, span)
			span.End()
		})
	}
}

func TestFailSpan(t *testing.T) {
	_, span := nooptrace.NewTracerProvider().Tracer("t").Start(context.Background(), "s")
	expectedErr := errors.New("something went wrong")

	got := tracing.FailSpan(span, expectedErr)
	assert.Equal(t, expectedErr, got)
	span.End()
}
