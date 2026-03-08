//go:build integration

package mongodb_test

import (
	"context"
	"testing"
	"time"

	"github.com/nawafswe/go-service-starter-kit/internal/clients/db/mongodb"
	nooptrace "go.opentelemetry.io/otel/trace/noop"
)

func TestNewClient_UnreachableHost(t *testing.T) {
	cfg := mongodb.Config{
		URI:            "mongodb://127.0.0.1:9999",
		ConnectTimeout: 500 * time.Millisecond,
	}
	_, err := mongodb.NewClient(context.Background(), cfg, nooptrace.NewTracerProvider())
	if err == nil {
		t.Fatal("expected error for unreachable host, got nil")
	}
}

func TestNewClient_InvalidURI(t *testing.T) {
	cfg := mongodb.Config{
		URI:            "not-a-valid-uri",
		ConnectTimeout: 500 * time.Millisecond,
	}
	_, err := mongodb.NewClient(context.Background(), cfg, nooptrace.NewTracerProvider())
	if err == nil {
		t.Fatal("expected error for invalid URI, got nil")
	}
}
