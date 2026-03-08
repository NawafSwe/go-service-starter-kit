package httperrors_test

import (
	"net/http"
	"testing"

	"github.com/nawafswe/go-service-starter-kit/internal/pkg/httperrors"
	"github.com/stretchr/testify/assert"
)

func TestBadRequestError(t *testing.T) {
	tests := []struct {
		name       string
		err        httperrors.BadRequestError
		wantError  string
		wantMsg    string
		wantCode   string
		wantStatus int
	}{
		{
			name:       "zero value",
			err:        httperrors.ErrBadRequest,
			wantError:  "bad request",
			wantMsg:    "bad request",
			wantCode:   "bad_request",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "with message",
			err:        httperrors.NewBadRequestError("invalid email"),
			wantError:  "bad request: invalid email",
			wantMsg:    "invalid email",
			wantCode:   "bad_request",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "with message and custom code",
			err:        httperrors.NewBadRequestErrorWithCode("too short", "validation_error"),
			wantError:  "bad request: too short",
			wantMsg:    "too short",
			wantCode:   "validation_error",
			wantStatus: http.StatusBadRequest,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.wantError, tc.err.Error())
			assert.Equal(t, tc.wantMsg, tc.err.Message())
			assert.Equal(t, tc.wantCode, tc.err.ErrorCode())
			assert.Equal(t, tc.wantStatus, tc.err.HTTPStatusCode())
		})
	}
}
