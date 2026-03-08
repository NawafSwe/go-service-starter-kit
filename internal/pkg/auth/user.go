package auth

import "github.com/nawafswe/go-service-starter-kit/internal/pkg/text"

// User is the authenticated principal extracted from a JWT token.
type User struct {
	ID             string
	Username       string
	DeviceID       string
	DeviceType     string
	AcceptLanguage string
	Email          string
	IsSubscribed   bool
	AuthToken      text.NonLoggable
}
