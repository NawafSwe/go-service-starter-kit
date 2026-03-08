package worker_test

import (
	"errors"
	"testing"

	"github.com/nawafswe/go-service-starter-kit/internal/observability/logger"
	"github.com/nawafswe/go-service-starter-kit/internal/worker"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestNewGRPCWorker(t *testing.T) {
	tests := []struct {
		name        string
		server      *grpc.Server
		logger      logger.Logger
		expectedErr error
	}{
		{name: "nil server", server: nil, logger: discardLogger(), expectedErr: errors.New("server is required to create grpc-worker")},
		{name: "nil logger", server: grpc.NewServer(), logger: nil, expectedErr: errors.New("logger is required to create grpc-worker")},
		{name: "valid", server: grpc.NewServer(), logger: discardLogger()},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w, err := worker.NewGRPCWorker(tc.server, tc.logger)
			if tc.expectedErr != nil {
				assert.EqualError(t, err, tc.expectedErr.Error())
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, w)
		})
	}
}
