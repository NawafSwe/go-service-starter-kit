package text_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/nawafswe/go-service-starter-kit/internal/text"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNonLoggable(t *testing.T) {
	tests := []struct {
		name string
		fn   func(nl text.NonLoggable) string
		want string
	}{
		{name: "String", fn: func(nl text.NonLoggable) string { return nl.String() }, want: "redacted"},
		{name: "GoString", fn: func(nl text.NonLoggable) string { return nl.GoString() }, want: "redacted"},
		{name: "GetValue", fn: func(nl text.NonLoggable) string { return nl.GetValue() }, want: "my-secret-token"},
		{name: "fmt %v", fn: func(nl text.NonLoggable) string { return fmt.Sprintf("%v", nl) }, want: "redacted"},
		{name: "fmt %#v", fn: func(nl text.NonLoggable) string { return fmt.Sprintf("%#v", nl) }, want: "redacted"},
		{
			name: "MarshalJSON",
			fn: func(nl text.NonLoggable) string {
				b, _ := json.Marshal(nl)
				return string(b)
			},
			want: `"redacted"`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			nl := text.NonLoggable("my-secret-token")
			assert.Equal(t, tc.want, tc.fn(nl))
		})
	}
}

func TestNonLoggable_InStruct(t *testing.T) {
	type Payload struct {
		Token text.NonLoggable `json:"token"`
	}
	p := Payload{Token: text.NonLoggable("secret")}
	b, err := json.Marshal(p)
	require.NoError(t, err)
	assert.JSONEq(t, `{"token":"redacted"}`, string(b))
}
