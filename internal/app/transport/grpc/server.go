// Package grpc provides the gRPC server transport for the application.
//
// # Getting started
//
//  1. Define your .proto files under docs/proto/.
//  2. Generate Go stubs:  protoc --go_out=. --go-grpc_out=. docs/proto/*.proto
//  3. Register your generated service implementations here in NewGRPCServer.
//  4. Add gRPC interceptors (auth, logging, tracing) as UnaryServerInterceptors.
//
// The example below shows the wiring skeleton — replace the placeholder with
// your actual service registration.
package grpc

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/auth"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/config"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/middleware"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/observability/logger"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// NewGRPCServer wires up and returns a configured *grpc.Server.
// Add your service registrations and interceptors here.
func NewGRPCServer(
	_ context.Context,
	cfg config.Config,
	_ *sqlx.DB,
	lgr logger.Logger,
) (*grpc.Server, error) {
	tracer := otel.Tracer("grpc.server")

	claimsParser, err := auth.NewClaimsParser(cfg.JWT.ISSUER, []byte(cfg.JWT.Secret))
	if err != nil {
		return nil, fmt.Errorf("grpc: failed to create claims parser: %w", err)
	}

	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			// Auth runs first — subsequent interceptors see the authenticated user via auth.UserFromCtx(ctx).
			// Swap GRPCAuthRequired for GRPCAuthOptional on endpoints that allow anonymous callers.
			middleware.GRPCAuthRequired(cfg.AuthorizationMock, claimsParser),
			middleware.GRPCLogging(lgr),
			middleware.GRPCTracing(tracer),
		),
	)

	// Register your gRPC services here, for example:
	//   pb.RegisterExampleServiceServer(srv, example.NewGRPCHandler(...))

	// Enable gRPC reflection for development tooling (grpcurl, etc.).
	if cfg.General.AppEnvironment != "production" {
		reflection.Register(srv)
	}

	return srv, nil
}
