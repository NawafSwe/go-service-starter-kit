package httperrors

import "net/http"

// ForbiddenError represents a 403 Forbidden transport error.
type ForbiddenError struct {
	message string
}

func (e ForbiddenError) Error() string {
	if e.message == "" {
		return "forbidden"
	}
	return "forbidden: " + e.message
}

func (e ForbiddenError) HTTPStatusCode() int { return http.StatusForbidden }

func (e ForbiddenError) Message() string {
	if e.message == "" {
		return "forbidden"
	}
	return e.message
}

func NewForbiddenError(message string) ForbiddenError {
	return ForbiddenError{message: message}
}
