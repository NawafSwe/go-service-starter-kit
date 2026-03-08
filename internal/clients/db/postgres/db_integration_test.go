//go:build integration

package postgres_test

import (
	"context"
	"testing"

	"github.com/nawafswe/go-service-starter-kit/internal/clients/db/postgres"
	"github.com/nawafswe/go-service-starter-kit/internal/config"
	nooptrace "go.opentelemetry.io/otel/trace/noop"
)

func TestNewConn_UnreachableHost(t *testing.T) {
	cfg := config.DB{
		DSN:                "postgres://user:pass@127.0.0.1:9999/testdb?sslmode=disable",
		MaxOpenConnections: 2,
		MaxIdleConnections: 1,
	}
	_, err := postgres.NewConn(context.Background(), cfg, "test", nooptrace.NewTracerProvider())
	if err == nil {
		t.Fatal("expected error for unreachable host, got nil")
	}
}
