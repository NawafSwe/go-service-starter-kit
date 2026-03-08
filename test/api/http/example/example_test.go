//go:build integration

package example_test

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/nawafswe/go-service-starter-kit/test/pkg/suite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/create_request.json
var createRequest []byte

func TestCreateExample(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		opts       func(t *testing.T) []suite.Option
		body       []byte
		headers    func(t *testing.T, s *suite.Suite) map[string]string
		assertFunc func(t *testing.T, s *suite.Suite, body []byte, status int)
	}{
		"creates example successfully": {
			body: createRequest,
			headers: func(t *testing.T, s *suite.Suite) map[string]string {
				return map[string]string{
					"Content-Type":  "application/json",
					"Authorization": "Bearer " + suite.MustGenerateToken(t, s, "user-1", "device-1"),
				}
			},
			assertFunc: func(t *testing.T, _ *suite.Suite, body []byte, status int) {
				assert.Equal(t, http.StatusCreated, status)

				var resp map[string]any
				require.NoError(t, json.Unmarshal(body, &resp))
				assert.Equal(t, "integration-test-example", resp["name"])
				assert.NotEmpty(t, resp["id"])
			},
		},
		"returns 401 without auth token": {
			body: createRequest,
			headers: func(_ *testing.T, _ *suite.Suite) map[string]string {
				return map[string]string{
					"Content-Type": "application/json",
				}
			},
			assertFunc: func(t *testing.T, _ *suite.Suite, _ []byte, status int) {
				assert.Equal(t, http.StatusUnauthorized, status)
			},
		},
		"returns 400 with empty name": {
			body: []byte(`{"name": ""}`),
			headers: func(t *testing.T, s *suite.Suite) map[string]string {
				return map[string]string{
					"Content-Type":  "application/json",
					"Authorization": "Bearer " + suite.MustGenerateToken(t, s, "user-1", "device-1"),
				}
			},
			assertFunc: func(t *testing.T, _ *suite.Suite, _ []byte, status int) {
				assert.Equal(t, http.StatusBadRequest, status)
			},
		},
		"returns 400 with invalid body": {
			body: []byte(`not-json`),
			headers: func(t *testing.T, s *suite.Suite) map[string]string {
				return map[string]string{
					"Content-Type":  "application/json",
					"Authorization": "Bearer " + suite.MustGenerateToken(t, s, "user-1", "device-1"),
				}
			},
			assertFunc: func(t *testing.T, _ *suite.Suite, _ []byte, status int) {
				assert.Equal(t, http.StatusBadRequest, status)
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var opts []suite.Option
			if tc.opts != nil {
				opts = tc.opts(t)
			}

			s := suite.SetupTestSuite(t, opts...)
			baseURL := suite.RunHTTPService(t, s)
			headers := tc.headers(t, s)

			body, status := suite.DoRequest(t, http.MethodPost, baseURL+"/app/api/v1/examples", bytes.NewReader(tc.body), headers)
			tc.assertFunc(t, s, body, status)
		})
	}
}

func TestGetExample(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		seedFunc   func(t *testing.T, s *suite.Suite, baseURL, token string) string
		headers    func(t *testing.T, s *suite.Suite) map[string]string
		assertFunc func(t *testing.T, body []byte, status int)
	}{
		"returns example by id": {
			seedFunc: func(t *testing.T, _ *suite.Suite, baseURL, token string) string {
				respBody, status := suite.DoRequest(t, http.MethodPost, baseURL+"/app/api/v1/examples", bytes.NewReader(createRequest), map[string]string{
					"Content-Type":  "application/json",
					"Authorization": "Bearer " + token,
				})
				require.Equal(t, http.StatusCreated, status)

				var created map[string]any
				require.NoError(t, json.Unmarshal(respBody, &created))
				return created["id"].(string)
			},
			headers: func(t *testing.T, s *suite.Suite) map[string]string {
				return map[string]string{
					"Authorization": "Bearer " + suite.MustGenerateToken(t, s, "user-1", "device-1"),
				}
			},
			assertFunc: func(t *testing.T, body []byte, status int) {
				assert.Equal(t, http.StatusOK, status)

				var resp map[string]any
				require.NoError(t, json.Unmarshal(body, &resp))
				assert.Equal(t, "integration-test-example", resp["name"])
			},
		},
		"returns 404 for non-existent id": {
			seedFunc: func(_ *testing.T, _ *suite.Suite, _, _ string) string {
				return "00000000-0000-0000-0000-000000000000"
			},
			headers: func(t *testing.T, s *suite.Suite) map[string]string {
				return map[string]string{
					"Authorization": "Bearer " + suite.MustGenerateToken(t, s, "user-1", "device-1"),
				}
			},
			assertFunc: func(t *testing.T, _ []byte, status int) {
				assert.Equal(t, http.StatusNotFound, status)
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			s := suite.SetupTestSuite(t)
			baseURL := suite.RunHTTPService(t, s)
			token := suite.MustGenerateToken(t, s, "user-1", "device-1")

			exampleID := tc.seedFunc(t, s, baseURL, token)
			headers := tc.headers(t, s)

			body, status := suite.DoRequest(t, http.MethodGet, fmt.Sprintf("%s/app/api/v1/examples/%s", baseURL, exampleID), nil, headers)
			tc.assertFunc(t, body, status)
		})
	}
}

func TestListExamples(t *testing.T) {
	t.Parallel()

	s := suite.SetupTestSuite(t)
	baseURL := suite.RunHTTPService(t, s)
	token := suite.MustGenerateToken(t, s, "user-1", "device-1")

	authHeaders := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + token,
	}

	// Seed two examples.
	for _, name := range []string{"example-a", "example-b"} {
		body, _ := json.Marshal(map[string]string{"name": name})
		_, status := suite.DoRequest(t, http.MethodPost, baseURL+"/app/api/v1/examples", bytes.NewReader(body), authHeaders)
		require.Equal(t, http.StatusCreated, status)
	}

	respBody, status := suite.DoRequest(t, http.MethodGet, baseURL+"/app/api/v1/examples", nil, authHeaders)
	assert.Equal(t, http.StatusOK, status)

	var items []map[string]any
	require.NoError(t, json.Unmarshal(respBody, &items))
	assert.Len(t, items, 2)
}

func TestDeleteExample(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		seedFunc   func(t *testing.T, baseURL, token string) string
		assertFunc func(t *testing.T, status int)
	}{
		"deletes example successfully": {
			seedFunc: func(t *testing.T, baseURL, token string) string {
				respBody, status := suite.DoRequest(t, http.MethodPost, baseURL+"/app/api/v1/examples", bytes.NewReader(createRequest), map[string]string{
					"Content-Type":  "application/json",
					"Authorization": "Bearer " + token,
				})
				require.Equal(t, http.StatusCreated, status)

				var created map[string]any
				require.NoError(t, json.Unmarshal(respBody, &created))
				return created["id"].(string)
			},
			assertFunc: func(t *testing.T, status int) {
				assert.Equal(t, http.StatusNoContent, status)
			},
		},
		"returns 404 for non-existent id": {
			seedFunc: func(_ *testing.T, _, _ string) string {
				return "00000000-0000-0000-0000-000000000000"
			},
			assertFunc: func(t *testing.T, status int) {
				assert.Equal(t, http.StatusNotFound, status)
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			s := suite.SetupTestSuite(t)
			baseURL := suite.RunHTTPService(t, s)
			token := suite.MustGenerateToken(t, s, "user-1", "device-1")

			exampleID := tc.seedFunc(t, baseURL, token)

			_, status := suite.DoRequest(t, http.MethodDelete, fmt.Sprintf("%s/app/api/v1/examples/%s", baseURL, exampleID), nil, map[string]string{
				"Authorization": "Bearer " + token,
			})
			tc.assertFunc(t, status)
		})
	}
}
