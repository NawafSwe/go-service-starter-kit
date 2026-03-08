package worker_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/nawafswe/go-service-starter-kit/internal/pkg/observability/logger"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/worker"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/worker/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConsumerWorker(t *testing.T) {
	ctrl := gomock.NewController(t)

	tests := []struct {
		name        string
		consumer    worker.MessageConsumer
		logger      logger.Logger
		expectedErr error
	}{
		{name: "nil consumer", consumer: nil, logger: discardLogger(), expectedErr: errors.New("consumer is required to create consumer-worker")},
		{name: "nil logger", consumer: mock.NewMockMessageConsumer(ctrl), logger: nil, expectedErr: errors.New("logger is required to create consumer-worker")},
		{name: "valid", consumer: mock.NewMockMessageConsumer(ctrl), logger: discardLogger()},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w, err := worker.NewConsumerWorker(tc.consumer, tc.logger)
			if tc.expectedErr != nil {
				assert.EqualError(t, err, tc.expectedErr.Error())
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, w)
		})
	}
}

func TestConsumerWorker_Run(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(m *mock.MockMessageConsumer)
		expectedErr error
	}{
		{
			name: "normal shutdown",
			setup: func(m *mock.MockMessageConsumer) {
				m.EXPECT().Start(gomock.Any()).Return(nil)
				m.EXPECT().Close().Return(nil)
			},
		},
		{
			name: "start error",
			setup: func(m *mock.MockMessageConsumer) {
				m.EXPECT().Start(gomock.Any()).Return(errors.New("broker unreachable"))
			},
			expectedErr: errors.New("broker unreachable"),
		},
		{
			name: "close error",
			setup: func(m *mock.MockMessageConsumer) {
				m.EXPECT().Start(gomock.Any()).Return(nil)
				m.EXPECT().Close().Return(errors.New("close failed"))
			},
			expectedErr: errors.New("close failed"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			m := mock.NewMockMessageConsumer(ctrl)
			tc.setup(m)

			w, _ := worker.NewConsumerWorker(m, discardLogger())

			done := make(chan error, 1)
			go func() { done <- w.Run(context.Background()) }()

			select {
			case err := <-done:
				if tc.expectedErr != nil {
					require.Error(t, err)
					assert.EqualError(t, err, tc.expectedErr.Error())
					return
				}
				assert.NoError(t, err)
			case <-time.After(2 * time.Second):
				t.Fatal("Run did not return within timeout")
			}
		})
	}
}
