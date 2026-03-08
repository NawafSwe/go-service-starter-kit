package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nawafswe/go-service-starter-kit/internal/auth"
	"github.com/nawafswe/go-service-starter-kit/internal/config"
	"github.com/nawafswe/go-service-starter-kit/internal/middleware"
)

func okHandler(t *testing.T) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func mustGenerateToken(t *testing.T, user auth.User) string {
	t.Helper()
	p, err := auth.NewClaimsParser(testIssuer, []byte(testSecret))
	if err != nil {
		t.Fatalf("NewClaimsParser: %v", err)
	}
	tok, err := p.GenerateJWTToken(user)
	if err != nil {
		t.Fatalf("GenerateJWTToken: %v", err)
	}
	return tok
}

func TestAuthRequired(t *testing.T) {
	p, _ := auth.NewClaimsParser(testIssuer, []byte(testSecret))
	validToken := mustGenerateToken(t, auth.User{ID: "u1"})

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
	}{
		{name: "valid token", authHeader: "Bearer " + validToken, wantStatus: http.StatusOK},
		{name: "missing header", authHeader: "", wantStatus: http.StatusUnauthorized},
		{name: "bad prefix", authHeader: "Token " + validToken, wantStatus: http.StatusUnauthorized},
		{name: "invalid token", authHeader: "Bearer garbage", wantStatus: http.StatusUnauthorized},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.AuthorizationMock{Enabled: false}
			handler := middleware.AuthRequired(cfg, p)(okHandler(t))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", rr.Code, tc.wantStatus)
			}
		})
	}
}

func TestAuthRequired_MockEnabled(t *testing.T) {
	p, _ := auth.NewClaimsParser(testIssuer, []byte(testSecret))
	cfg := config.AuthorizationMock{Enabled: true, UserID: "mock-user", DeviceID: "mock-device"}

	var gotUser auth.User
	handler := middleware.AuthRequired(cfg, p)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUser = auth.UserFromCtx(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if gotUser.ID != "mock-user" {
		t.Errorf("user ID = %q, want %q", gotUser.ID, "mock-user")
	}
}

func TestAuthOptional(t *testing.T) {
	p, _ := auth.NewClaimsParser(testIssuer, []byte(testSecret))
	validToken := mustGenerateToken(t, auth.User{ID: "u1"})

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
		wantUser   bool
	}{
		{name: "no header passes through", authHeader: "", wantStatus: http.StatusOK, wantUser: false},
		{name: "valid token injects user", authHeader: "Bearer " + validToken, wantStatus: http.StatusOK, wantUser: true},
		{name: "invalid token returns 401", authHeader: "Bearer garbage", wantStatus: http.StatusUnauthorized, wantUser: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.AuthorizationMock{Enabled: false}
			var gotUser auth.User
			inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotUser = auth.UserFromCtx(r.Context())
				w.WriteHeader(http.StatusOK)
			})
			handler := middleware.AuthOptional(cfg, p)(inner)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", rr.Code, tc.wantStatus)
			}
			if tc.wantUser && gotUser.ID == "" {
				t.Error("expected user injected into context, got empty")
			}
			if !tc.wantUser && gotUser.ID != "" {
				t.Errorf("expected no user, got ID %q", gotUser.ID)
			}
		})
	}
}

func TestAuthMock(t *testing.T) {
	tests := []struct {
		name     string
		enabled  bool
		wantUser string
	}{
		{name: "enabled injects mock user", enabled: true, wantUser: "mock-id"},
		{name: "disabled passes through", enabled: false, wantUser: ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.AuthorizationMock{Enabled: tc.enabled, UserID: "mock-id"}
			var gotUser auth.User
			inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotUser = auth.UserFromCtx(r.Context())
				w.WriteHeader(http.StatusOK)
			})
			handler := middleware.AuthMock(cfg)(inner)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if gotUser.ID != tc.wantUser {
				t.Errorf("user ID = %q, want %q", gotUser.ID, tc.wantUser)
			}
		})
	}
}
