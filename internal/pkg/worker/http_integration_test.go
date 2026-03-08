//go:build integration

package worker_test

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/nawafswe/go-service-starter-kit/internal/pkg/worker"
)

func TestHTTPWorker_Run_Shutdown(t *testing.T) {
	addr := freeAddr(t)
	srv := &http.Server{Addr: addr}
	w, _ := worker.NewHTTPWorker(srv, discardLogger())

	done := make(chan error, 1)
	go func() { done <- w.Run(context.Background()) }()

	time.Sleep(30 * time.Millisecond)

	// Close triggers ListenAndServe → http.ErrServerClosed (filtered) → errCh
	// closed → Run proceeds to Shutdown (no-op on closed server) → returns nil.
	srv.Close()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run returned unexpected error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return within timeout")
	}
}

func TestHTTPWorker_Run_ListenError(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("setup listen: %v", err)
	}
	defer l.Close()

	srv := &http.Server{Addr: l.Addr().String()}
	w, _ := worker.NewHTTPWorker(srv, discardLogger())

	done := make(chan error, 1)
	go func() { done <- w.Run(context.Background()) }()

	select {
	case err := <-done:
		if err == nil {
			t.Error("expected error when port is occupied, got nil")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return within timeout")
	}
}
