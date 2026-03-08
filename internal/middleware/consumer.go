package middleware

import (
	"context"
	"fmt"

	"github.com/nawafswe/go-service-starter-kit/internal/auth"
	"github.com/nawafswe/go-service-starter-kit/internal/config"
)

// ConsumerAuthRequired validates a JWT token string extracted from a message header/payload
// and injects the authenticated user into the context.
// Returns an error when the token is missing, expired, or invalid.
//
// Example — Kafka message processing:
//
//	token := msg.Headers["Authorization"]  // or from msg payload
//	ctx, err = middleware.ConsumerAuthRequired(claimsParser)(ctx, token)
func ConsumerAuthRequired(claimsParser auth.ClaimsParser) func(ctx context.Context, token string) (context.Context, error) {
	return func(ctx context.Context, token string) (context.Context, error) {
		if token == "" {
			return ctx, fmt.Errorf("consumer auth: missing jwt token in message")
		}
		user, err := claimsParser.ParseJWTToken(token)
		if err != nil {
			return ctx, fmt.Errorf("consumer auth: invalid or expired token: %w", err)
		}
		return auth.SetUserCtx(ctx, user), nil
	}
}

// ConsumerAuthOptional is like ConsumerAuthRequired but allows messages without a token.
// If the token is empty the context is returned unchanged — user.ID will be empty.
// If a token is present but invalid, it returns an error.
func ConsumerAuthOptional(claimsParser auth.ClaimsParser) func(ctx context.Context, token string) (context.Context, error) {
	return func(ctx context.Context, token string) (context.Context, error) {
		if token == "" {
			return ctx, nil
		}
		user, err := claimsParser.ParseJWTToken(token)
		if err != nil {
			return ctx, fmt.Errorf("consumer auth: invalid or expired token: %w", err)
		}
		return auth.SetUserCtx(ctx, user), nil
	}
}

// ConsumerAuthMock injects a fixed user from config — for local development only.
// Never enable in production.
func ConsumerAuthMock(cfg config.AuthorizationMock) func(ctx context.Context, _ string) (context.Context, error) {
	return func(ctx context.Context, _ string) (context.Context, error) {
		if !cfg.Enabled {
			return ctx, nil
		}
		mockUser := auth.User{ID: cfg.UserID, DeviceID: cfg.DeviceID}
		return auth.SetUserCtx(ctx, mockUser), nil
	}
}
