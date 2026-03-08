package auth_test

import (
	"errors"
	"testing"
	"time"

	"github.com/nawafswe/go-service-starter-kit/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testIssuer = "test-issuer"
	testSecret = "super-secret-key-for-testing"
)

func defaultParser(t *testing.T) auth.ClaimsParser {
	t.Helper()
	p, err := auth.NewClaimsParser(testIssuer, []byte(testSecret))
	require.NoError(t, err)
	return p
}

func TestNewClaimsParser(t *testing.T) {
	tests := []struct {
		name        string
		issuer      string
		secret      []byte
		opts        []auth.ClaimsParserOption
		expectedErr error
	}{
		{name: "valid", issuer: testIssuer, secret: []byte(testSecret)},
		{name: "valid with TTL option", issuer: testIssuer, secret: []byte(testSecret), opts: []auth.ClaimsParserOption{auth.WithClaimsParserTTL(time.Hour)}},
		{name: "missing issuer", issuer: "", secret: []byte(testSecret), expectedErr: errors.New("jwt: issuer is required")},
		{name: "nil secret", issuer: testIssuer, secret: nil, expectedErr: errors.New("jwt: secret is required")},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := auth.NewClaimsParser(tc.issuer, tc.secret, tc.opts...)
			if tc.expectedErr != nil {
				assert.EqualError(t, err, tc.expectedErr.Error())
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestGenerateAndParseJWTToken(t *testing.T) {
	tests := []struct {
		name        string
		genIssuer   string
		parseIssuer string
		user        auth.User
		expectedErr error
	}{
		{
			name:        "valid round trip",
			genIssuer:   testIssuer,
			parseIssuer: testIssuer,
			user:        auth.User{ID: "user-1", Username: "testuser", Email: "test@example.com", DeviceID: "device-1", DeviceType: "mobile", IsSubscribed: true},
		},
		{
			name:        "wrong issuer",
			genIssuer:   "other-issuer",
			parseIssuer: testIssuer,
			user:        auth.User{ID: "u1"},
			expectedErr: errors.New("failed to parse jwt token"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			genParser, _ := auth.NewClaimsParser(tc.genIssuer, []byte(testSecret))
			parseParser, _ := auth.NewClaimsParser(tc.parseIssuer, []byte(testSecret))

			token, err := genParser.GenerateJWTToken(tc.user)
			require.NoError(t, err)

			got, err := parseParser.ParseJWTToken(token)
			if tc.expectedErr != nil {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr.Error())
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.user.ID, got.ID)
			assert.Equal(t, tc.user.Username, got.Username)
			assert.Equal(t, tc.user.Email, got.Email)
			assert.Equal(t, tc.user.IsSubscribed, got.IsSubscribed)
		})
	}
}

func TestParseJWTToken_Invalid(t *testing.T) {
	p := defaultParser(t)

	tests := []struct {
		name        string
		token       string
		expectedErr error
	}{
		{name: "empty token", token: "", expectedErr: errors.New("failed to parse jwt token")},
		{name: "garbage", token: "not.a.token", expectedErr: errors.New("failed to parse jwt token")},
		{name: "wrong signature", token: "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJ4In0.wrong", expectedErr: errors.New("failed to parse jwt token")},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := p.ParseJWTToken(tc.token)
			require.Error(t, err)
			assert.ErrorContains(t, err, tc.expectedErr.Error())
		})
	}
}
