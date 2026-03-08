package httperrors

import "net/http"

// BadRequestError represents a 400 Bad Request transport error.
type BadRequestError struct {
	message string
	code    string
}

func (e BadRequestError) Error() string {
	if e.message == "" {
		return "bad request"
	}
	return "bad request: " + e.message
}

func (e BadRequestError) HTTPStatusCode() int { return http.StatusBadRequest }

func (e BadRequestError) Message() string {
	if e.message == "" {
		return "bad request"
	}
	return e.message
}

func (e BadRequestError) ErrorCode() string {
	if e.code != "" {
		return e.code
	}
	return "bad_request"
}

var ErrBadRequest = BadRequestError{}

func NewBadRequestError(message string) BadRequestError {
	return BadRequestError{message: message, code: "bad_request"}
}

func NewBadRequestErrorWithCode(message, code string) BadRequestError {
	return BadRequestError{message: message, code: code}
}
