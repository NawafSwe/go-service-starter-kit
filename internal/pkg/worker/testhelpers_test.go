package worker_test

import (
	"io"
	"net"
	"testing"

	"github.com/nawafswe/go-service-starter-kit/internal/pkg/observability/logger"
)

func discardLogger() logger.ZerologLogger {
	return logger.NewLogger(logger.ErrorLevel, "test", "0.0.0", "test").WithOutput(io.Discard)
}

func freeAddr(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("freeAddr: %v", err)
	}
	addr := l.Addr().String()
	l.Close()
	return addr
}
