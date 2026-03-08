//go:build integration

package mysql_test

import (
	"context"
	"testing"

	"github.com/nawafswe/go-service-starter-kit/internal/clients/db/mysql"
	"github.com/nawafswe/go-service-starter-kit/internal/config"
	"github.com/stretchr/testify/assert"
	nooptrace "go.opentelemetry.io/otel/trace/noop"
)

func TestNewConn_UnreachableHost(t *testing.T) {
	cfg := config.DB{
		DSN:                "user:pass@tcp(127.0.0.1:9999)/testdb",
		MaxOpenConnections: 2,
		MaxIdleConnections: 1,
	}

	tests := []struct {
		name string
		opts []mysql.Option
	}{
		{
			name: "without options",
		},
		{
			name: "with tracer provider",
			opts: []mysql.Option{mysql.WithTracerProvider(nooptrace.NewTracerProvider())},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := mysql.NewConn(context.Background(), cfg, "test", tc.opts...)
			assert.Error(t, err)
		})
	}
}
