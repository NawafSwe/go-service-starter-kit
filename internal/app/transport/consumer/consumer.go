// Package consumer provides the message consumer transport.
//
// # Getting started
//
//  1. Choose a broker client (e.g. github.com/segmentio/kafka-go, github.com/IBM/sarama).
//  2. Implement the broker subscription loop in Start().
//  3. Wire JWT auth via middleware.ConsumerAuthRequired / ConsumerAuthOptional when
//     the message carries a user token in its headers.
//
// Each consumed message is decoded and routed through go-kit endpoints with the
// same middleware chain (timeout, rate-limit, logging) as the HTTP and gRPC transports.
package consumer

import (
	"context"
	"fmt"

	"github.com/nawafswe/go-service-starter-kit/internal/app/transport/consumer/bootstrap"
	v1 "github.com/nawafswe/go-service-starter-kit/internal/app/transport/consumer/v1"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/auth"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/config"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/middleware"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/observability/logger"
)

// Consumer implements worker.MessageConsumer by routing decoded messages
// through go-kit endpoints. Replace the Start/Close stubs with your broker client.
type Consumer struct {
	cfg            config.Config
	lgr            logger.Logger
	endpoints      *bootstrap.MessageEndpoints
	authMiddleware func(ctx context.Context, token string) (context.Context, error)
}

// NewConsumer wires up the consumer with its dependencies.
func NewConsumer(
	_ context.Context,
	cfg config.Config,
	deps *bootstrap.Dependencies,
	resources *bootstrap.SharedResource,
) (*Consumer, error) {
	repos, err := bootstrap.InitializeRepositories(cfg, deps)
	if err != nil {
		return nil, fmt.Errorf("consumer: failed to initialize repositories: %w", err)
	}

	endpoints := bootstrap.InitializeEndpoints(cfg, repos)

	claimsParser, err := auth.NewClaimsParser(cfg.JWT.ISSUER, []byte(cfg.JWT.Secret))
	if err != nil {
		return nil, fmt.Errorf("consumer: failed to create claims parser: %w", err)
	}

	return &Consumer{
		cfg:       cfg,
		lgr:       resources.Lgr,
		endpoints: endpoints,
		// Use ConsumerAuthRequired when your messages always carry a user token.
		// Use ConsumerAuthOptional when the token is optional.
		// Use ConsumerAuthMock for local development when auth.mock is enabled.
		authMiddleware: chooseAuthMiddleware(cfg, claimsParser),
	}, nil
}

// Start begins consuming messages.
// This method blocks until ctx is done or an unrecoverable error occurs.
// Replace the TODO with your actual broker subscription loop.
func (c *Consumer) Start(ctx context.Context) error {
	c.lgr.Info(ctx, "consumer started, waiting for messages...")

	// TODO: initialise your broker client and subscribe to cfg.Consumer.Topics
	// Example with kafka-go:
	//
	//   reader := kafka.NewReader(kafka.ReaderConfig{
	//       Brokers: c.cfg.Consumer.Brokers,
	//       GroupID: c.cfg.Consumer.GroupID,
	//       Topic:   c.cfg.Consumer.Topics[0],
	//   })
	//   defer reader.Close()
	//
	//   for {
	//       msg, err := reader.ReadMessage(ctx)
	//       if err != nil {
	//           if errors.Is(err, context.Canceled) { return nil }
	//           return fmt.Errorf("consumer: read message: %w", err)
	//       }
	//       if err := c.handleMessage(ctx, msg.Headers["Authorization"], msg.Value); err != nil {
	//           c.lgr.Error(ctx, err, "failed to handle message")
	//       }
	//   }

	<-ctx.Done()
	return nil
}

// Close performs a graceful shutdown.
func (c *Consumer) Close() error {
	// TODO: close your broker client here.
	return nil
}

// handleMessage authenticates and routes a single message through the go-kit endpoint chain.
func (c *Consumer) handleMessage(ctx context.Context, msgType string, token string, payload []byte) error {
	// Step 1: authenticate the message producer (optional — remove if not needed).
	ctx, err := c.authMiddleware(ctx, token)
	if err != nil {
		return fmt.Errorf("consumer: auth failed: %w", err)
	}

	// Step 2: route to the appropriate endpoint based on message type.
	switch msgType {
	case "example.create":
		req, err := v1.DecodeCreateExampleCommand(payload)
		if err != nil {
			return err
		}
		_, err = c.endpoints.CreateExample(ctx, req)
		return err

	default:
		c.lgr.InfoFields(ctx, "unknown message type", map[string]any{"type": msgType})
		return nil
	}
}

// chooseAuthMiddleware selects the right JWT middleware based on config.
func chooseAuthMiddleware(cfg config.Config, cp auth.ClaimsParser) func(ctx context.Context, token string) (context.Context, error) {
	if cfg.AuthorizationMock.Enabled {
		return middleware.ConsumerAuthMock(cfg.AuthorizationMock)
	}
	return middleware.ConsumerAuthOptional(cp)
}
