package domain

import "errors"

// Sentinel errors used across service and repository layers.
// The transport layer maps these to HTTP / gRPC status codes.
var (
	ErrNotFound       = errors.New("not_found")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrInvalidRequest = errors.New("invalid_request")
	ErrConflict       = errors.New("conflict")
	ErrForbidden      = errors.New("forbidden")
)
