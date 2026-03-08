package httperrors_test

import (
	"net/http"
	"testing"

	"github.com/nawafswe/go-service-starter-kit/internal/pkg/httperrors"
	"github.com/stretchr/testify/assert"
)

func TestNotFoundError(t *testing.T) {
	tests := []struct {
		name       string
		err        httperrors.NotFoundError
		wantError  string
		wantMsg    string
		wantStatus int
	}{
		{
			name:       "with name",
			err:        httperrors.NotFoundError{Name: "user"},
			wantError:  "user not found",
			wantMsg:    "user not found",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "empty name",
			err:        httperrors.NotFoundError{},
			wantError:  " not found",
			wantMsg:    " not found",
			wantStatus: http.StatusNotFound,
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
