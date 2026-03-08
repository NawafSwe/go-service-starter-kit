package httperrors_test

import (
	"net/http"
	"testing"

	"github.com/nawafswe/go-service-starter-kit/internal/httperrors"
	"github.com/stretchr/testify/assert"
)

func TestUnauthorizedError(t *testing.T) {
	tests := []struct {
		name       string
		err        httperrors.UnauthorizedError
		wantError  string
		wantMsg    string
		wantStatus int
	}{
		{
			name:       "zero value",
			err:        httperrors.UnauthorizedError{},
			wantError:  "unauthorized",
			wantMsg:    "unauthorized",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "with message",
			err:        httperrors.NewUnauthorizedError("token expired"),
			wantError:  "unauthorized: token expired",
			wantMsg:    "token expired",
			wantStatus: http.StatusUnauthorized,
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
