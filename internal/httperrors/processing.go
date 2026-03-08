package httperrors

import (
	"fmt"
	"net/http"
)

// ProcessingError wraps an internal error with a 500 status code.
type ProcessingError struct {
	Name string
	Err  error
}

func (e ProcessingError) Error() string {
	return fmt.Sprintf("error processing %s: %s", e.Name, e.Err.Error())
}
func (e ProcessingError) Message() string     { return "internal server error" }
func (e ProcessingError) HTTPStatusCode() int { return http.StatusInternalServerError }
func (e ProcessingError) Unwrap() error       { return e.Err }
