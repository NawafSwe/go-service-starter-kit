package httperrors_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/nawafswe/go-service-starter-kit/internal/httperrors"
	"github.com/stretchr/testify/assert"
)

func TestProcessingError(t *testing.T) {
	dbErr := errors.New("connection refused")

	tests := []struct {
		name       string
		err        httperrors.ProcessingError
		wantError  string
		wantMsg    string
		wantStatus int
		wantUnwrap error
	}{
		{
			name:       "wraps internal error",
			err:        httperrors.ProcessingError{Name: "create user", Err: dbErr},
			wantError:  "error processing create user: connection refused",
			wantMsg:    "internal server error",
			wantStatus: http.StatusInternalServerError,
			wantUnwrap: dbErr,
		},
		{
			name:       "different operation name",
			err:        httperrors.ProcessingError{Name: "fetch orders", Err: errors.New("timeout")},
			wantError:  "error processing fetch orders: timeout",
			wantMsg:    "internal server error",
			wantStatus: http.StatusInternalServerError,
			wantUnwrap: errors.New("timeout"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.wantError, tc.err.Error())
			assert.Equal(t, tc.wantMsg, tc.err.Message())
			assert.Equal(t, tc.wantStatus, tc.err.HTTPStatusCode())
			assert.EqualError(t, tc.err.Unwrap(), tc.wantUnwrap.Error())
		})
	}
}

func TestProcessingError_ErrorsIs(t *testing.T) {
	inner := errors.New("connection refused")
	err := httperrors.ProcessingError{Name: "save", Err: inner}
	assert.ErrorIs(t, err, inner)
}
