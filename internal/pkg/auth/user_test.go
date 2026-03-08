package auth_test

import (
	"context"
	"testing"

	"github.com/nawafswe/go-service-starter-kit/internal/pkg/auth"
	"github.com/stretchr/testify/assert"
)

func TestUserFromCtx(t *testing.T) {
	tests := []struct {
		name     string
		user     *auth.User
		wantID   string
		wantName string
	}{
		{
			name:     "set and retrieve user",
			user:     &auth.User{ID: "u1", Username: "alice"},
			wantID:   "u1",
			wantName: "alice",
		},
		{
			name:   "empty context returns zero user",
			user:   nil,
			wantID: "",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			if tc.user != nil {
				ctx = auth.SetUserCtx(ctx, *tc.user)
			}

			got := auth.UserFromCtx(ctx)
			assert.Equal(t, tc.wantID, got.ID)
			assert.Equal(t, tc.wantName, got.Username)
		})
	}
}
