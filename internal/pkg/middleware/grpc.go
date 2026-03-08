package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/nawafswe/go-service-starter-kit/internal/pkg/auth"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/config"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/observability/logger"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// GRPCAuthRequired validates the JWT on every inbound RPC call.
// Returns codes.Unauthenticated when the metadata header is missing,
// malformed, expired, or has an invalid signature.
// When cfg.Enabled is true the mock interceptor is used instead.
func GRPCAuthRequired(cfg config.AuthorizationMock, claimsParser auth.ClaimsParser) grpc.UnaryServerInterceptor {
	if cfg.Enabled {
		return GRPCAuthMock(cfg)
	}
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		token, err := grpcBearerToken(ctx)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		user, err := claimsParser.ParseJWTToken(token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
		}
		return handler(auth.SetUserCtx(ctx, user), req)
	}
}

// GRPCAuthOptional validates the JWT only when the authorization metadata key is present.
// RPCs without the header continue unauthenticated — user.ID will be empty.
// When cfg.Enabled is true the mock interceptor is used instead.
func GRPCAuthOptional(cfg config.AuthorizationMock, claimsParser auth.ClaimsParser) grpc.UnaryServerInterceptor {
	if cfg.Enabled {
		return GRPCAuthMock(cfg)
	}
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok || len(md.Get(grpcAuthHeader)) == 0 {
			return handler(ctx, req)
		}
		token, err := grpcBearerToken(ctx)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		user, err := claimsParser.ParseJWTToken(token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
		}
		return handler(auth.SetUserCtx(ctx, user), req)
	}
}

// GRPCAuthMock injects a fixed user from config — for local development only.
// Never enable in production.
func GRPCAuthMock(cfg config.AuthorizationMock) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if !cfg.Enabled {
			return handler(ctx, req)
		}
		mockUser := auth.User{ID: cfg.UserID, DeviceID: cfg.DeviceID}
		return handler(auth.SetUserCtx(ctx, mockUser), req)
	}
}

// GRPCLogging returns a unary server interceptor that logs every inbound RPC call
// and any errors the handler returns.
func GRPCLogging(lgr logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		lgr.InfoW(ctx, fmt.Sprintf("gRPC call: %s", info.FullMethod), map[string]any{
			"method": info.FullMethod,
		})
		resp, err := handler(ctx, req)
		if err != nil {
			lgr.Error(ctx, err, fmt.Sprintf("gRPC error: %s", info.FullMethod))
		}
		return resp, err
	}
}

// GRPCTracing returns a unary server interceptor that starts an OTel span for every call.
func GRPCTracing(tracer trace.Tracer) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		ctx, span := tracer.Start(ctx, info.FullMethod)
		defer span.End()
		return handler(ctx, req)
	}
}

// grpcAuthHeader is the canonical gRPC metadata key for the authorization header.
const grpcAuthHeader = "authorization"

// grpcBearerToken extracts and validates the Bearer token from incoming gRPC metadata.
func grpcBearerToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "missing metadata")
	}
	values := md.Get(grpcAuthHeader)
	if len(values) == 0 {
		return "", status.Error(codes.Unauthenticated, "authorization header is missing")
	}
	raw := values[0]
	if !strings.HasPrefix(raw, bearerPrefix) {
		return "", status.Error(codes.Unauthenticated, "authorization header must start with 'Bearer '")
	}
	token := strings.TrimPrefix(raw, bearerPrefix)
	if token == "" {
		return "", status.Error(codes.Unauthenticated, "bearer token is empty")
	}
	return token, nil
}
