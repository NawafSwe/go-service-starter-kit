//go:build integration

package mysql_test

import (
	"context"
	"testing"

	nooptrace "go.opentelemetry.io/otel/trace/noop"

	"github.com/nawafswe/go-service-starter-kit/internal/pkg/clients/db/mysql"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/config"
)

func TestNewConn_UnreachableHost(t *testing.T) {
	cfg := config.DB{
		DSN:                "user:pass@tcp(127.0.0.1:9999)/testdb",
		MaxOpenConnections: 2,
		MaxIdleConnections: 1,
	}
	_, err := mysql.NewConn(context.Background(), cfg, "test", nooptrace.NewTracerProvider())
	if err == nil {
		t.Fatal("expected error for unreachable host, got nil")
	}
}
