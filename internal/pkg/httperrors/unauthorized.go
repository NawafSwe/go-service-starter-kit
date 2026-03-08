package httperrors

import "net/http"

// UnauthorizedError represents a 401 Unauthorized transport error.
type UnauthorizedError struct {
	message string
}

func (e UnauthorizedError) Error() string {
	if e.message == "" {
		return "unauthorized"
	}
	return "unauthorized: " + e.message
}

func (e UnauthorizedError) HTTPStatusCode() int { return http.StatusUnauthorized }

func (e UnauthorizedError) Message() string {
	if e.message == "" {
		return "unauthorized"
	}
	return e.message
}

func NewUnauthorizedError(message string) UnauthorizedError {
	return UnauthorizedError{message: message}
}
