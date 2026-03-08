package auth_test

import (
	"errors"
	"testing"

	"github.com/nawafswe/go-service-starter-kit/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasswordGenerator(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		compare     string
		expectedErr error
	}{
		{name: "encrypt and compare correct password", password: "s3cr3tP@ssw0rd", compare: "s3cr3tP@ssw0rd"},
		{name: "compare wrong password", password: "correct", compare: "wrong", expectedErr: errors.New("invalid credentials")},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pg := auth.NewPasswordGenerator()
			hashed, err := pg.EncryptPassword(tc.password)
			require.NoError(t, err)
			assert.NotEmpty(t, hashed)
			assert.NotEqual(t, tc.password, hashed)

			err = pg.ComparePassword(hashed, tc.compare)
			if tc.expectedErr != nil {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr.Error())
				return
			}
			assert.NoError(t, err)
		})
	}
}
