package worker

//go:generate mockgen -destination=mock/mock.go -package=mock github.com/nawafswe/go-service-starter-kit/internal/worker MessageConsumer

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/nawafswe/go-service-starter-kit/internal/observability/logger"
)

const ConsumerWorkerName = "consumer-worker"

// MessageConsumer is the interface that any message consumer must implement.
// Implementations are responsible for connecting to the broker, subscribing to
// topics, and processing messages until the context is cancelled.
type MessageConsumer interface {
	// Start begins consuming messages. It must block until ctx is done or an
	// unrecoverable error occurs.
	Start(ctx context.Context) error
	// Close performs a graceful shutdown, flushing any in-flight processing.
	Close() error
}

// ConsumerWorker runs a MessageConsumer with graceful shutdown on SIGINT/SIGTERM.
type ConsumerWorker struct {
	consumer MessageConsumer
	lgr      logger.Logger
}

func NewConsumerWorker(consumer MessageConsumer, lgr logger.Logger) (*ConsumerWorker, error) {
	if consumer == nil {
		return nil, fmt.Errorf("consumer is required to create %s", ConsumerWorkerName)
	}
	if lgr == nil {
		return nil, fmt.Errorf("logger is required to create %s", ConsumerWorkerName)
	}
	return &ConsumerWorker{consumer: consumer, lgr: lgr}, nil
}

func (c *ConsumerWorker) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(stop)

	errCh := make(chan error, 1)

	c.lgr.Info(ctx, "consumer worker starting...")
	go func() {
		if err := c.consumer.Start(ctx); err != nil {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case sig := <-stop:
		c.lgr.Info(ctx, fmt.Sprintf("shutdown signal received: %s", sig))
		cancel()
	case err := <-errCh:
		if err != nil {
			c.lgr.Error(ctx, err, "consumer stopped with error")
			return err
		}
	}

	c.lgr.Info(ctx, "closing consumer...")
	if err := c.consumer.Close(); err != nil {
		c.lgr.Error(ctx, err, "error closing consumer")
		return err
	}
	c.lgr.Info(ctx, "consumer shutdown complete")
	return nil
}
