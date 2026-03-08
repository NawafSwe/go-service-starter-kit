// Package grpc provides the gRPC server transport for the application.
//
// # Getting started
//
//  1. Define your .proto files under docs/api/proto/.
//  2. Generate Go stubs:  protoc --go_out=. --go-grpc_out=. docs/proto/*.proto
//  3. Register your generated service implementations in NewGRPCServer.
//  4. The standard interceptor chain (auth, logging, tracing) is applied via middleware.
//
// Each RPC method is wired through go-kit endpoints with the same middleware
// chain (timeout, rate-limit, logging) as the HTTP transport.
package grpc

import (
	"context"
	"fmt"

	grpcv1 "github.com/nawafswe/go-service-starter-kit/api/proto/grpc/v1/gen"
	"github.com/nawafswe/go-service-starter-kit/internal/app/transport/grpc/bootstrap"
	"github.com/nawafswe/go-service-starter-kit/internal/auth"
	"github.com/nawafswe/go-service-starter-kit/internal/config"
	"github.com/nawafswe/go-service-starter-kit/internal/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// NewGRPCServer builds the gRPC server — initialises repositories, wires handlers,
// and returns a configured *grpc.Server ready to be handed to the gRPC worker.
func NewGRPCServer(
	_ context.Context,
	cfg config.Config,
	deps *bootstrap.Dependencies,
	resources *bootstrap.SharedResource,
) (*grpc.Server, error) {
	repos, err := bootstrap.InitializeRepositories(cfg, deps)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize repositories: %w", err)
	}

	claimsParser, err := auth.NewClaimsParser(cfg.JWT.ISSUER, []byte(cfg.JWT.Secret))
	if err != nil {
		return nil, fmt.Errorf("grpc: failed to create claims parser: %w", err)
	}

	interceptors := []grpc.UnaryServerInterceptor{
		middleware.GRPCAuthRequired(cfg.AuthorizationMock, claimsParser),
		middleware.GRPCLogging(resources.Lgr),
	}
	if resources.Tracer != nil {
		interceptors = append(interceptors, middleware.GRPCTracing(resources.Tracer))
	}

	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(interceptors...),
	)

	// Register example service — wired through go-kit endpoints.
	exampleHandler := bootstrap.InitializeServiceHandler(cfg, resources, repos)
	grpcv1.RegisterExampleServiceServer(srv, exampleHandler)

	// Enable gRPC reflection for development tooling (grpcurl, etc.).
	if cfg.General.AppEnvironment != "production" {
		reflection.Register(srv)
	}

	return srv, nil
}
