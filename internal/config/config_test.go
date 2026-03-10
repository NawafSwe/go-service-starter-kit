package config_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/nawafswe/go-service-starter-kit/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatalf("writeFile %s: %v", name, err)
	}
	return p
}

const validYAML = `
GENERAL:
  SERVICE_NAME: test-service
  APP_VERSION: "0.0.1"
  APP_ENVIRONMENT: development
  LOG_LEVEL: debug
  TRACING:
    ENABLED: false

HTTP:
  PORT: 8080
  SHUTDOWN_TIMEOUT: 15s
  READ_TIMEOUT: 5s
  READ_HEADER_TIMEOUT: 2s
  WRITE_TIMEOUT: 10s

GRPC:
  PORT: 50051

CONSUMER:
  BROKERS: []
  GROUP_ID: ""
  TOPICS: []

DB:
  DSN: "postgres://localhost/testdb"
  MAX_OPEN_CONNECTIONS: 10
  MAX_IDLE_CONNECTIONS: 5
  MAX_CONNECTIONS_LIFETIME: 30m

JWT:
  SECRET: "test-secret"
  TTL: 24h
  ISSUER: "test-issuer"

METRICS:
  ENABLED: false

AUTHORIZATION_MOCK:
  ENABLED: false
  USER_ID: ""
  DEVICE_ID: ""
`

const minimalEnv = "# test env\n"

func TestLoad(t *testing.T) {
	tests := []struct {
		name         string
		yaml         string
		env          string
		osEnv        map[string]string
		skipYAML     bool
		yamlPathFunc func(dir string) string
		skipEnv      bool
		expectedErr  error
		checkCfg     func(t *testing.T, cfg config.Config)
	}{
		{
			name: "valid config",
			yaml: validYAML,
			env:  minimalEnv,
			checkCfg: func(t *testing.T, cfg config.Config) {
				assert.Equal(t, "test-secret", cfg.JWT.Secret)
				assert.Equal(t, 8080, cfg.HTTP.Port)
				assert.Equal(t, "postgres://localhost/testdb", cfg.DB.DSN)
			},
		},
		{
			name: "env file overrides YAML",
			yaml: validYAML,
			env:  "JWT__SECRET=overridden-secret\n",
			checkCfg: func(t *testing.T, cfg config.Config) {
				assert.Equal(t, "overridden-secret", cfg.JWT.Secret)
			},
		},
		{
			name:  "OS env overrides env file",
			yaml:  validYAML,
			env:   "JWT__SECRET=from-env-file\n",
			osEnv: map[string]string{"JWT__SECRET": "from-os-env"},
			checkCfg: func(t *testing.T, cfg config.Config) {
				assert.Equal(t, "from-os-env", cfg.JWT.Secret)
			},
		},
		{
			name:         "invalid YAML path",
			env:          minimalEnv,
			yamlPathFunc: func(dir string) string { return filepath.Join(dir, "nonexistent.yaml") },
			expectedErr:  errors.New("config: read yaml"),
		},
		{
			name:     "env only without YAML",
			env:      "JWT__SECRET=env-secret\nDB__DSN=postgres://localhost/testdb\nHTTP__PORT=9090\n",
			skipYAML: true,
			checkCfg: func(t *testing.T, cfg config.Config) {
				assert.Equal(t, "env-secret", cfg.JWT.Secret)
				assert.Equal(t, "postgres://localhost/testdb", cfg.DB.DSN)
				assert.Equal(t, 9090, cfg.HTTP.Port)
			},
		},
		{
			name:        "missing env file",
			yaml:        validYAML,
			env:         minimalEnv,
			skipEnv:     true,
			expectedErr: errors.New("config: read env"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()

			var yamlPath string
			switch {
			case tc.yamlPathFunc != nil:
				yamlPath = tc.yamlPathFunc(dir)
			case !tc.skipYAML:
				yamlPath = writeFile(t, dir, "config.yaml", tc.yaml)
			}

			envPath := filepath.Join(dir, "nonexistent.env")
			if !tc.skipEnv {
				envPath = writeFile(t, dir, ".env", tc.env)
			}

			for k, v := range tc.osEnv {
				t.Setenv(k, v)
			}

			cfg, err := config.Load(yamlPath, envPath)
			if tc.expectedErr != nil {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr.Error())
				return
			}
			require.NoError(t, err)
			if tc.checkCfg != nil {
				tc.checkCfg(t, cfg)
			}
		})
	}
}

func TestLoad_Validate(t *testing.T) {
	tests := []struct {
		name        string
		yaml        string
		expectedErr error
	}{
		{
			name: "missing JWT secret",
			yaml: `
GENERAL:
  APP_ENVIRONMENT: development
HTTP:
  PORT: 8080
DB:
  DSN: "postgres://localhost/testdb"
JWT:
  SECRET: ""
`,
			expectedErr: errors.New("config: JWT__SECRET is required"),
		},
		{
			name: "missing DB DSN",
			yaml: `
GENERAL:
  APP_ENVIRONMENT: development
HTTP:
  PORT: 8080
JWT:
  SECRET: "s"
DB:
  DSN: ""
`,
			expectedErr: errors.New("config: DB__DSN is required"),
		},
		{
			name: "missing HTTP port",
			yaml: `
GENERAL:
  APP_ENVIRONMENT: development
HTTP:
  PORT: 0
JWT:
  SECRET: "s"
DB:
  DSN: "postgres://localhost/testdb"
`,
			expectedErr: errors.New("config: HTTP__PORT is required"),
		},
		{
			name: "mock enabled in production",
			yaml: `
GENERAL:
  APP_ENVIRONMENT: production
HTTP:
  PORT: 8080
JWT:
  SECRET: "s"
DB:
  DSN: "postgres://localhost/testdb"
AUTHORIZATION_MOCK:
  ENABLED: true
`,
			expectedErr: errors.New("config: AUTHORIZATION_MOCK__ENABLED must not be true in production"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			yamlPath := writeFile(t, dir, "config.yaml", tc.yaml)
			envPath := writeFile(t, dir, ".env", minimalEnv)

			_, err := config.Load(yamlPath, envPath)
			assert.EqualError(t, err, tc.expectedErr.Error())
		})
	}
}
