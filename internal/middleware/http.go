package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/nawafswe/go-service-starter-kit/internal/auth"
	"github.com/nawafswe/go-service-starter-kit/internal/config"
)

const (
	authorizationHeader  = "Authorization"
	bearerPrefix         = "Bearer "
	acceptLanguageHeader = "Accept-Language"

	enLang = "en"
	arLang = "ar"
)

// AuthRequired validates the JWT on every request.
// Returns 401 if the header is missing, malformed, expired, or has an invalid signature.
//
// Flow:
//  1. Extract Bearer token from Authorization header
//  2. Parse and validate token (signature + expiry)
//  3. Inject authenticated User into context
//  4. Call next — or short-circuit with 401
func AuthRequired(cfg config.AuthorizationMock, claimsParser auth.ClaimsParser) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if cfg.Enabled {
			return AuthMock(cfg)(next)
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr, err := extractBearerToken(r)
			if err != nil {
				writeUnauthorized(w, "missing or malformed authorization header")
				return
			}
			user, err := claimsParser.ParseJWTToken(tokenStr)
			if err != nil {
				writeUnauthorized(w, "invalid or expired token")
				return
			}
			user.AcceptLanguage = extractLanguage(r)
			next.ServeHTTP(w, r.WithContext(auth.SetUserCtx(r.Context(), user)))
		})
	}
}

// AuthOptional validates the JWT only when the Authorization header is present.
// Requests without the header continue unauthenticated — user.ID will be empty.
// Use on endpoints that behave differently for authenticated vs anonymous callers.
func AuthOptional(cfg config.AuthorizationMock, claimsParser auth.ClaimsParser) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if cfg.Enabled {
			return AuthMock(cfg)(next)
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get(authorizationHeader) == "" {
				next.ServeHTTP(w, r)
				return
			}
			tokenStr, err := extractBearerToken(r)
			if err != nil {
				writeUnauthorized(w, "missing or malformed authorization header")
				return
			}
			user, err := claimsParser.ParseJWTToken(tokenStr)
			if err != nil {
				writeUnauthorized(w, "invalid or expired token")
				return
			}
			user.AcceptLanguage = extractLanguage(r)
			next.ServeHTTP(w, r.WithContext(auth.SetUserCtx(r.Context(), user)))
		})
	}
}

// AuthMock bypasses JWT validation and injects a fixed user from config.
// Only active when cfg.Enabled is true — never enable in production.
func AuthMock(cfg config.AuthorizationMock) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !cfg.Enabled {
				next.ServeHTTP(w, r)
				return
			}
			mockUser := auth.User{ID: cfg.UserID, DeviceID: cfg.DeviceID}
			next.ServeHTTP(w, r.WithContext(auth.SetUserCtx(r.Context(), mockUser)))
		})
	}
}

func writeUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = fmt.Fprintf(w, `{"message":%q,"error":"unauthorized"}`, message)
}

func extractBearerToken(r *http.Request) (string, error) {
	header := r.Header.Get(authorizationHeader)
	if header == "" {
		return "", fmt.Errorf("authorization header is missing")
	}
	if !strings.HasPrefix(header, bearerPrefix) {
		return "", fmt.Errorf("authorization header must start with 'Bearer '")
	}
	token := strings.TrimPrefix(header, bearerPrefix)
	if token == "" {
		return "", fmt.Errorf("bearer token is empty")
	}
	return token, nil
}

func extractLanguage(r *http.Request) string {
	lng := r.Header.Get(acceptLanguageHeader)
	if strings.HasPrefix(lng, enLang) {
		return enLang
	}
	return arLang
}
