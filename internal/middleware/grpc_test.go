package middleware_test

import (
	"context"
	"testing"

	"github.com/nawafswe/go-service-starter-kit/internal/auth"
	"github.com/nawafswe/go-service-starter-kit/internal/config"
	"github.com/nawafswe/go-service-starter-kit/internal/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func incomingCtx(authHeader string) context.Context {
	md := metadata.Pairs("authorization", authHeader)
	return metadata.NewIncomingContext(context.Background(), md)
}

func noopHandler(_ context.Context, req any) (any, error) {
	return req, nil
}

func TestGRPCAuthRequired(t *testing.T) {
	p := defaultClaimsParser(t)
	validToken := generateToken(t, auth.User{ID: "u1"})

	tests := []struct {
		name     string
		ctx      context.Context
		wantCode codes.Code
		wantUser string
	}{
		{
			name:     "valid token",
			ctx:      incomingCtx("Bearer " + validToken),
			wantCode: codes.OK,
			wantUser: "u1",
		},
		{
			name:     "missing metadata",
			ctx:      context.Background(),
			wantCode: codes.Unauthenticated,
		},
		{
			name:     "invalid token",
			ctx:      incomingCtx("Bearer garbage"),
			wantCode: codes.Unauthenticated,
		},
		{
			name:     "bad prefix",
			ctx:      incomingCtx("Token " + validToken),
			wantCode: codes.Unauthenticated,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			interceptor := middleware.GRPCAuthRequired(config.AuthorizationMock{}, p)
			var gotUser auth.User
			handler := func(ctx context.Context, req any) (any, error) {
				gotUser = auth.UserFromCtx(ctx)
				return nil, nil
			}
			_, err := interceptor(tc.ctx, nil, &grpc.UnaryServerInfo{}, handler)
			st, _ := status.FromError(err)
			if st.Code() != tc.wantCode {
				t.Errorf("code = %v, want %v", st.Code(), tc.wantCode)
			}
			if tc.wantUser != "" && gotUser.ID != tc.wantUser {
				t.Errorf("user ID = %q, want %q", gotUser.ID, tc.wantUser)
			}
		})
	}
}

func TestGRPCAuthRequired_MockEnabled(t *testing.T) {
	p := defaultClaimsParser(t)
	cfg := config.AuthorizationMock{Enabled: true, UserID: "mock-id", DeviceID: "dev-1"}
	interceptor := middleware.GRPCAuthRequired(cfg, p)

	var gotUser auth.User
	handler := func(ctx context.Context, req any) (any, error) {
		gotUser = auth.UserFromCtx(ctx)
		return nil, nil
	}
	_, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{}, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotUser.ID != "mock-id" {
		t.Errorf("user ID = %q, want %q", gotUser.ID, "mock-id")
	}
}

func TestGRPCAuthOptional(t *testing.T) {
	p := defaultClaimsParser(t)
	validToken := generateToken(t, auth.User{ID: "u2"})

	tests := []struct {
		name     string
		ctx      context.Context
		wantCode codes.Code
		wantUser string
	}{
		{
			name:     "no metadata passes through",
			ctx:      context.Background(),
			wantCode: codes.OK,
			wantUser: "",
		},
		{
			name:     "valid token injects user",
			ctx:      incomingCtx("Bearer " + validToken),
			wantCode: codes.OK,
			wantUser: "u2",
		},
		{
			name:     "invalid token returns unauthenticated",
			ctx:      incomingCtx("Bearer garbage"),
			wantCode: codes.Unauthenticated,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			interceptor := middleware.GRPCAuthOptional(config.AuthorizationMock{}, p)
			var gotUser auth.User
			handler := func(ctx context.Context, req any) (any, error) {
				gotUser = auth.UserFromCtx(ctx)
				return nil, nil
			}
			_, err := interceptor(tc.ctx, nil, &grpc.UnaryServerInfo{}, handler)
			st, _ := status.FromError(err)
			if st.Code() != tc.wantCode {
				t.Errorf("code = %v, want %v", st.Code(), tc.wantCode)
			}
			if tc.wantUser != "" && gotUser.ID != tc.wantUser {
				t.Errorf("user ID = %q, want %q", gotUser.ID, tc.wantUser)
			}
		})
	}
}

func TestGRPCAuthMock(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
		wantID  string
	}{
		{name: "enabled injects mock user", enabled: true, wantID: "mock-id"},
		{name: "disabled passes through", enabled: false, wantID: ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.AuthorizationMock{Enabled: tc.enabled, UserID: "mock-id"}
			interceptor := middleware.GRPCAuthMock(cfg)
			var gotUser auth.User
			handler := func(ctx context.Context, req any) (any, error) {
				gotUser = auth.UserFromCtx(ctx)
				return nil, nil
			}
			_, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{}, handler)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotUser.ID != tc.wantID {
				t.Errorf("user ID = %q, want %q", gotUser.ID, tc.wantID)
			}
		})
	}
}
