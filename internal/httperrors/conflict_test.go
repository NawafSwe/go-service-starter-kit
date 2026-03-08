package httperrors_test

import (
	"net/http"
	"testing"

	"github.com/nawafswe/go-service-starter-kit/internal/httperrors"
	"github.com/stretchr/testify/assert"
)

func TestConflictError(t *testing.T) {
	tests := []struct {
		name       string
		err        httperrors.ConflictError
		wantError  string
		wantMsg    string
		wantStatus int
	}{
		{
			name:       "zero value",
			err:        httperrors.ConflictError{},
			wantError:  "conflict",
			wantMsg:    "conflict",
			wantStatus: http.StatusConflict,
		},
		{
			name:       "with message",
			err:        httperrors.NewConflictError("email already exists"),
			wantError:  "conflict: email already exists",
			wantMsg:    "email already exists",
			wantStatus: http.StatusConflict,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.wantError, tc.err.Error())
			assert.Equal(t, tc.wantMsg, tc.err.Message())
			assert.Equal(t, tc.wantStatus, tc.err.HTTPStatusCode())
		})
	}
}
