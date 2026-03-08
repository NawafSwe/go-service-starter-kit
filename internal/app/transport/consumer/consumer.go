// Package consumer provides the message consumer transport.
//
// # Getting started
//
//  1. Choose a broker client (e.g. github.com/segmentio/kafka-go, github.com/IBM/sarama).
//  2. Implement the business logic handler for each topic.
//  3. Wire JWT auth via middleware.ConsumerAuthRequired / ConsumerAuthOptional when
//     the message carries a user token in its headers.
//
// The Consumer in this file is a stub that demonstrates the wiring pattern.
// Replace the TODO comments with your actual broker client and message handlers.
package consumer

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/auth"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/config"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/middleware"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/observability/logger"
)

// Consumer is the stub message consumer.
// Replace the fields and Start/Close implementations with your broker client.
type Consumer struct {
	cfg            config.Config
	lgr            logger.Logger
	db             *sqlx.DB
	authMiddleware func(ctx context.Context, token string) (context.Context, error)
}

// NewConsumer wires up the consumer with its dependencies.
func NewConsumer(
	_ context.Context,
	cfg config.Config,
	db *sqlx.DB,
	lgr logger.Logger,
) (*Consumer, error) {
	claimsParser, err := auth.NewClaimsParser(cfg.JWT.ISSUER, []byte(cfg.JWT.Secret))
	if err != nil {
		return nil, fmt.Errorf("consumer: failed to create claims parser: %w", err)
	}

	return &Consumer{
		cfg: cfg,
		lgr: lgr,
		db:  db,
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
	//       if err := c.handleMessage(ctx, msg); err != nil {
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

// handleMessage processes a single message.
// It validates the JWT from the message header before calling the business handler.
//
// Wire your actual business handlers here following the same pattern as the
// HTTP transport: endpoint → business handler → repository.
func (c *Consumer) handleMessage(ctx context.Context, token string, payload []byte) error {
	// Step 1: authenticate the message producer (optional — remove if not needed).
	ctx, err := c.authMiddleware(ctx, token)
	if err != nil {
		return fmt.Errorf("consumer: auth failed: %w", err)
	}

	// Step 2: route to the appropriate business handler based on message type / topic.
	// TODO: replace this with your actual message routing and handler calls.
	c.lgr.InfoFields(ctx, "handling message", map[string]any{"payload_len": len(payload)})
	return nil
}

// chooseAuthMiddleware selects the right JWT middleware based on config.
func chooseAuthMiddleware(cfg config.Config, cp auth.ClaimsParser) func(ctx context.Context, token string) (context.Context, error) {
	if cfg.AuthorizationMock.Enabled {
		return middleware.ConsumerAuthMock(cfg.AuthorizationMock)
	}
	return middleware.ConsumerAuthOptional(cp)
}
