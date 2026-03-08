//go:build integration

package tracing_test

import (
	"context"
	"testing"
	"time"

	"github.com/nawafswe/go-service-starter-kit/internal/config"
	"github.com/nawafswe/go-service-starter-kit/internal/observability/tracing"
)

func TestSetup_Enabled(t *testing.T) {
	cfg := config.Config{}
	cfg.General.Tracing.Enabled = true
	cfg.General.Tracing.ReceiverEndpoint = "127.0.0.1:9999" // no collector running
	cfg.General.ServiceName = "test-svc"
	cfg.General.AppVersion = "0.0.1"
	cfg.General.AppEnvironment = "test"

	// grpc.NewClient is non-blocking and exporter creation is lazy,
	// so Setup succeeds even without a running collector.
	tp, shutdown, err := tracing.Setup(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Setup: %v", err)
	}
	if tp == nil {
		t.Fatal("expected non-nil tracer provider")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	_ = shutdown(ctx) // flush errors are acceptable with no running collector
}
