//go:build integration

package worker_test

import (
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"

	"github.com/nawafswe/go-service-starter-kit/internal/pkg/worker"
)

func TestGRPCWorker_Run_Shutdown(t *testing.T) {
	addr := freeAddr(t)
	srv := grpc.NewServer()
	w, _ := worker.NewGRPCWorker(srv, discardLogger())
	w.WithAddr(addr)

	done := make(chan error, 1)
	go func() { done <- w.Run(context.Background()) }()

	time.Sleep(30 * time.Millisecond)

	// GracefulStop causes srv.Serve to return nil → errCh closes → Run returns nil.
	srv.GracefulStop()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run returned unexpected error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return within timeout")
	}
}

func TestGRPCWorker_Run_ListenError(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("setup listen: %v", err)
	}
	defer l.Close()

	w, _ := worker.NewGRPCWorker(grpc.NewServer(), discardLogger())
	w.WithAddr(l.Addr().String())

	err = w.Run(context.Background())
	if err == nil {
		t.Error("expected error when port is occupied, got nil")
	}
}
