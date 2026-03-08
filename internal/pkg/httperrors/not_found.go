package httperrors

import (
	"fmt"
	"net/http"
)

// NotFoundError represents a 404 Not Found transport error.
type NotFoundError struct {
	Name string
}

func (e NotFoundError) Error() string       { return fmt.Sprintf("%s not found", e.Name) }
func (e NotFoundError) HTTPStatusCode() int { return http.StatusNotFound }
func (e NotFoundError) Message() string     { return fmt.Sprintf("%s not found", e.Name) }
