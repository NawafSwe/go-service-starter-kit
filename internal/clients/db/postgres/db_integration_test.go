//go:build integration

package postgres_test

import (
	"context"
	"testing"

	"github.com/nawafswe/go-service-starter-kit/internal/clients/db/postgres"
	"github.com/nawafswe/go-service-starter-kit/internal/config"
	"github.com/stretchr/testify/assert"
	nooptrace "go.opentelemetry.io/otel/trace/noop"
)

func TestNewConn_UnreachableHost(t *testing.T) {
	cfg := config.DB{
		DSN:                "postgres://user:pass@127.0.0.1:9999/testdb?sslmode=disable",
		MaxOpenConnections: 2,
		MaxIdleConnections: 1,
	}

	tests := []struct {
		name string
		opts []postgres.Option
	}{
		{
			name: "without options",
		},
		{
			name: "with tracer provider",
			opts: []postgres.Option{postgres.WithTracerProvider(nooptrace.NewTracerProvider())},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := postgres.NewConn(context.Background(), cfg, "test", tc.opts...)
			assert.Error(t, err)
		})
	}
}
