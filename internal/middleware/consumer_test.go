package middleware_test

import (
	"context"
	"testing"

	"github.com/nawafswe/go-service-starter-kit/internal/auth"
	"github.com/nawafswe/go-service-starter-kit/internal/config"
	"github.com/nawafswe/go-service-starter-kit/internal/middleware"
)

const (
	testIssuer = "test-issuer"
	testSecret = "super-secret-key-for-testing"
)

func defaultClaimsParser(t *testing.T) auth.ClaimsParser {
	t.Helper()
	p, err := auth.NewClaimsParser(testIssuer, []byte(testSecret))
	if err != nil {
		t.Fatalf("NewClaimsParser: %v", err)
	}
	return p
}

func generateToken(t *testing.T, user auth.User) string {
	t.Helper()
	p := defaultClaimsParser(t)
	tok, err := p.GenerateJWTToken(user)
	if err != nil {
		t.Fatalf("GenerateJWTToken: %v", err)
	}
	return tok
}

func TestConsumerAuthRequired(t *testing.T) {
	p := defaultClaimsParser(t)
	validToken := generateToken(t, auth.User{ID: "u1", Username: "alice"})

	tests := []struct {
		name    string
		token   string
		wantErr bool
		wantID  string
	}{
		{name: "valid token", token: validToken, wantErr: false, wantID: "u1"},
		{name: "empty token", token: "", wantErr: true},
		{name: "invalid token", token: "not-a-token", wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fn := middleware.ConsumerAuthRequired(p)
			ctx, err := fn(context.Background(), tc.token)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := auth.UserFromCtx(ctx)
			if got.ID != tc.wantID {
				t.Errorf("user ID = %q, want %q", got.ID, tc.wantID)
			}
		})
	}
}

func TestConsumerAuthOptional(t *testing.T) {
	p := defaultClaimsParser(t)
	validToken := generateToken(t, auth.User{ID: "u2"})

	tests := []struct {
		name    string
		token   string
		wantErr bool
		wantID  string
	}{
		{name: "empty token passes through", token: "", wantErr: false, wantID: ""},
		{name: "valid token injects user", token: validToken, wantErr: false, wantID: "u2"},
		{name: "invalid token returns error", token: "garbage", wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fn := middleware.ConsumerAuthOptional(p)
			ctx, err := fn(context.Background(), tc.token)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := auth.UserFromCtx(ctx)
			if got.ID != tc.wantID {
				t.Errorf("user ID = %q, want %q", got.ID, tc.wantID)
			}
		})
	}
}

func TestConsumerAuthMock(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
		wantID  string
	}{
		{name: "enabled injects mock user", enabled: true, wantID: "mock-user"},
		{name: "disabled passes through", enabled: false, wantID: ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.AuthorizationMock{Enabled: tc.enabled, UserID: "mock-user"}
			fn := middleware.ConsumerAuthMock(cfg)
			ctx, err := fn(context.Background(), "")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := auth.UserFromCtx(ctx)
			if got.ID != tc.wantID {
				t.Errorf("user ID = %q, want %q", got.ID, tc.wantID)
			}
		})
	}
}
