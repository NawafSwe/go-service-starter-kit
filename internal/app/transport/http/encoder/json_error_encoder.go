package encoder

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/jsonapi"
	"github.com/nawafswe/go-service-starter-kit/internal/app/domain"
	transporterrors "github.com/nawafswe/go-service-starter-kit/internal/httperrors"
)

const (
	contextCancelledStatusCode = 499
	contextCancelledError      = "request is cancelled"
)

type (
	errorHTTPStatusCoder interface {
		HTTPStatusCode() int
	}
	errorMessenger interface {
		Message() string
	}
)

// JSONErrorEncoder encodes domain and transport errors as JSON:API error responses.
type JSONErrorEncoder struct{}

func NewJSONErrorEncoder() JSONErrorEncoder { return JSONErrorEncoder{} }

func (enc JSONErrorEncoder) Encode(ctx context.Context, err error, w http.ResponseWriter) {
	code := http.StatusInternalServerError
	message := "internal server error"

	if isCanceledError(ctx, err) {
		code = contextCancelledStatusCode
		message = contextCancelledError
	} else {
		code, message = domainErrorCode(err)

		var errHTTPStatusCoder errorHTTPStatusCoder
		if errors.As(err, &errHTTPStatusCoder) {
			code = errHTTPStatusCoder.HTTPStatusCode()
		}
		var errMessenger errorMessenger
		if errors.As(err, &errMessenger) {
			message = errMessenger.Message()
		}
	}

	w.Header().Set("Content-Type", jsonapi.MediaType)
	w.WriteHeader(code)
	_ = jsonapi.MarshalErrors(w, []*jsonapi.ErrorObject{{Title: message}})
}

func domainErrorCode(err error) (int, string) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, "not found"
	case errors.Is(err, domain.ErrUnauthorized):
		return http.StatusUnauthorized, "unauthorized"
	case errors.Is(err, domain.ErrInvalidRequest):
		return http.StatusBadRequest, "invalid request"
	case errors.Is(err, domain.ErrConflict):
		return http.StatusConflict, "conflict"
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden, "forbidden"
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}

func isCanceledError(ctx context.Context, err error) bool {
	if errors.Is(ctx.Err(), context.Canceled) || errors.Is(err, context.Canceled) {
		return true
	}
	var processingErr *transporterrors.ProcessingError
	if errors.As(err, &processingErr) {
		return errors.Is(processingErr.Err, context.Canceled)
	}
	return false
}
