package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nawafswe/go-service-starter-kit/internal/pkg/text"
)

const defaultTTL = 24 * time.Hour

// jwtCtxKey is the unexported context key for injected users.
type jwtCtxKey struct{}

// Claims holds the JWT payload.
// Standard claims (exp, iss, sub) are embedded and validated automatically.
type Claims struct {
	jwt.RegisteredClaims
	UserID       string `json:"user_id"`
	DeviceID     string `json:"device_id"`
	DeviceType   string `json:"device_type"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	IsSubscribed bool   `json:"is_subscribed"`
}

// ClaimsParserOption is a functional option for ClaimsParser.
type ClaimsParserOption func(*ClaimsParser)

// WithClaimsParserTTL overrides the default token TTL.
func WithClaimsParserTTL(ttl time.Duration) ClaimsParserOption {
	return func(c *ClaimsParser) {
		if ttl > 0 {
			c.ttl = ttl
		}
	}
}

// ClaimsParser handles JWT generation and validation.
type ClaimsParser struct {
	ttl    time.Duration
	issuer string
	secret []byte
}

// NewClaimsParser creates a ClaimsParser. Both issuer and secret are required.
func NewClaimsParser(issuer string, secret []byte, opts ...ClaimsParserOption) (ClaimsParser, error) {
	if issuer == "" {
		return ClaimsParser{}, fmt.Errorf("jwt: issuer is required")
	}
	if secret == nil {
		return ClaimsParser{}, fmt.Errorf("jwt: secret is required")
	}
	c := ClaimsParser{ttl: defaultTTL, secret: secret, issuer: issuer}
	for _, opt := range opts {
		opt(&c)
	}
	return c, nil
}

// ParseJWTToken validates the token string and extracts the user.
// Returns an error if the token is expired, has an invalid signature, or is malformed.
func (c ClaimsParser) ParseJWTToken(token string) (User, error) {
	jwtToken, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (any, error) {
		// Reject tokens signed with unexpected algorithms — prevents "alg: none" attack.
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return c.secret, nil
	}, jwt.WithIssuer(c.issuer), jwt.WithExpirationRequired())
	if err != nil {
		return User{}, fmt.Errorf("failed to parse jwt token: %w", err)
	}
	claims, ok := jwtToken.Claims.(*Claims)
	if !ok || !jwtToken.Valid {
		return User{}, fmt.Errorf("invalid jwt token")
	}
	return User{
		ID:           claims.UserID,
		Username:     claims.Username,
		AuthToken:    text.NonLoggable(token),
		DeviceID:     claims.DeviceID,
		DeviceType:   claims.DeviceType,
		Email:        claims.Email,
		IsSubscribed: claims.IsSubscribed,
	}, nil
}

// GenerateJWTToken creates and signs a JWT for the given user.
func (c ClaimsParser) GenerateJWTToken(user User) (string, error) {
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(c.ttl)),
			Issuer:    c.issuer,
		},
		UserID:       user.ID,
		Username:     user.Username,
		Email:        user.Email,
		DeviceID:     user.DeviceID,
		DeviceType:   user.DeviceType,
		IsSubscribed: user.IsSubscribed,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(c.secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign jwt token: %w", err)
	}
	return signed, nil
}

// SetUserCtx injects the authenticated user into the request context.
func SetUserCtx(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, jwtCtxKey{}, user)
}

// UserFromCtx extracts the authenticated user from context.
// Returns an empty User{} if not present — callers should check user.ID != "".
func UserFromCtx(ctx context.Context) User {
	user, _ := ctx.Value(jwtCtxKey{}).(User)
	return user
}
