package httperrors_test

import (
	"net/http"
	"testing"

	"github.com/nawafswe/go-service-starter-kit/internal/pkg/httperrors"
	"github.com/stretchr/testify/assert"
)

func TestForbiddenError(t *testing.T) {
	tests := []struct {
		name       string
		err        httperrors.ForbiddenError
		wantError  string
		wantMsg    string
		wantStatus int
	}{
		{
			name:       "zero value",
			err:        httperrors.ForbiddenError{},
			wantError:  "forbidden",
			wantMsg:    "forbidden",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "with message",
			err:        httperrors.NewForbiddenError("insufficient permissions"),
			wantError:  "forbidden: insufficient permissions",
			wantMsg:    "insufficient permissions",
			wantStatus: http.StatusForbidden,
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
