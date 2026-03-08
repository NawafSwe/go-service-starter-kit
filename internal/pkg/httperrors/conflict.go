package httperrors

import "net/http"

// ConflictError represents a 409 Conflict transport error.
type ConflictError struct {
	message string
}

func (e ConflictError) Error() string {
	if e.message == "" {
		return "conflict"
	}
	return "conflict: " + e.message
}

func (e ConflictError) HTTPStatusCode() int { return http.StatusConflict }

func (e ConflictError) Message() string {
	if e.message == "" {
		return "conflict"
	}
	return e.message
}

func NewConflictError(message string) ConflictError {
	return ConflictError{message: message}
}
