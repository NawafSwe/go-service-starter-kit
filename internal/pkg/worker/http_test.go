package worker_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/nawafswe/go-service-starter-kit/internal/pkg/observability/logger"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/worker"
	"github.com/stretchr/testify/assert"
)

func TestNewHTTPWorker(t *testing.T) {
	tests := []struct {
		name        string
		server      *http.Server
		logger      logger.Logger
		expectedErr error
	}{
		{name: "nil server", server: nil, logger: discardLogger(), expectedErr: errors.New("server is required to create http-worker")},
		{name: "nil logger", server: &http.Server{}, logger: nil, expectedErr: errors.New("logger is required to create http-worker")},
		{name: "valid", server: &http.Server{}, logger: discardLogger()},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w, err := worker.NewHTTPWorker(tc.server, tc.logger)
			if tc.expectedErr != nil {
				assert.EqualError(t, err, tc.expectedErr.Error())
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, w)
		})
	}
}
