// Package config loads service configuration from a YAML file and a .env file,
// merging them with environment variables. The .env file takes highest priority.
package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/nawafswe/go-service-starter-kit/internal/grpcx"
	"github.com/nawafswe/go-service-starter-kit/internal/httpx"
	"github.com/spf13/viper"
)

const ServiceName = "go-backend-kit-starter"

// Config is the full application configuration.
type Config struct {
	General           General           `mapstructure:"GENERAL"`
	HTTP              HTTP              `mapstructure:"HTTP"`
	GRPC              GRPC              `mapstructure:"GRPC"`
	Consumer          Consumer          `mapstructure:"CONSUMER"`
	DB                DB                `mapstructure:"DB"`
	JWT               JWT               `mapstructure:"JWT"`
	Metrics           Metrics           `mapstructure:"METRICS"`
	AuthorizationMock AuthorizationMock `mapstructure:"AUTHORIZATION_MOCK"`
	Endpoints         Endpoints         `mapstructure:"ENDPOINTS"`
	Clients           Clients           `mapstructure:"CLIENTS"`
}

type Clients struct {
	HTTP map[string]httpx.Config `mapstructure:"HTTP"`
	GRPC map[string]grpcx.Config `mapstructure:"GRPC"`
}

// AuthorizationMock bypasses JWT validation for local development.
// NEVER enable in production.
type AuthorizationMock struct {
	Enabled  bool   `mapstructure:"ENABLED"`
	UserID   string `mapstructure:"USER_ID"`
	DeviceID string `mapstructure:"DEVICE_ID"`
}

type General struct {
	ServiceName    string  `mapstructure:"SERVICE_NAME"`
	AppVersion     string  `mapstructure:"APP_VERSION"`
	AppEnvironment string  `mapstructure:"APP_ENVIRONMENT"`
	LogLevel       string  `mapstructure:"LOG_LEVEL"`
	Tracing        Tracing `mapstructure:"TRACING"`
}

type HTTP struct {
	Port              int           `mapstructure:"PORT"`
	ShutdownTimeout   time.Duration `mapstructure:"SHUTDOWN_TIMEOUT"`
	ReadTimeout       time.Duration `mapstructure:"READ_TIMEOUT"`
	ReadHeaderTimeout time.Duration `mapstructure:"READ_HEADER_TIMEOUT"`
	WriteTimeout      time.Duration `mapstructure:"WRITE_TIMEOUT"`
}

type GRPC struct {
	Port            int           `mapstructure:"PORT"`
	ShutdownTimeout time.Duration `mapstructure:"SHUTDOWN_TIMEOUT"`
}

// Consumer holds configuration for message broker consumers (Kafka, RabbitMQ, etc.).
type Consumer struct {
	Brokers []string `mapstructure:"BROKERS"`
	GroupID string   `mapstructure:"GROUP_ID"`
	Topics  []string `mapstructure:"TOPICS"`
}

type DB struct {
	DSN                    string        `mapstructure:"DSN"`
	MaxOpenConnections     int           `mapstructure:"MAX_OPEN_CONNECTIONS"`
	MaxIdleConnections     int           `mapstructure:"MAX_IDLE_CONNECTIONS"`
	MaxConnectionsLifetime time.Duration `mapstructure:"MAX_CONNECTIONS_LIFETIME"`
}

type JWT struct {
	Secret string        `mapstructure:"SECRET"`
	TTL    time.Duration `mapstructure:"TTL"`
	ISSUER string        `mapstructure:"ISSUER"`
}

type Tracing struct {
	Enabled          bool   `mapstructure:"ENABLED"`
	ReceiverEndpoint string `mapstructure:"RECEIVER_ENDPOINT"`
}

type Metrics struct {
	Enabled bool `mapstructure:"ENABLED"`
}

// Endpoints holds per-endpoint rate limiter and timeout configuration.
type Endpoints struct {
	ExampleCreate EndpointCfg `mapstructure:"EXAMPLE_CREATE"`
	ExampleGet    EndpointCfg `mapstructure:"EXAMPLE_GET"`
	ExampleList   EndpointCfg `mapstructure:"EXAMPLE_LIST"`
	ExampleDelete EndpointCfg `mapstructure:"EXAMPLE_DELETE"`
}

// EndpointCfg holds per-endpoint rate limiter and deadline settings.
type EndpointCfg struct {
	Deadline    time.Duration  `mapstructure:"DEADLINE"`
	RateLimiter RateLimiterCfg `mapstructure:"RATE_LIMITER"`
}

// RateLimiterCfg is a sliding-window rate limiter configuration.
type RateLimiterCfg struct {
	Interval time.Duration `mapstructure:"INTERVAL"`
	Limit    int64         `mapstructure:"LIMIT"`
}

// Load reads configuration from an optional YAML file and a .env file, merging them.
// If configPath is empty, YAML loading is skipped and only .env + OS environment are used.
// Priority: OS environment > .env file > YAML defaults.
func Load(configPath, envPath string) (Config, error) {
	const delimiter = "__"

	base := viper.NewWithOptions(viper.KeyDelimiter(delimiter))

	if configPath != "" {
		base.SetConfigType("yaml")
		base.SetConfigFile(configPath)
		if err := base.ReadInConfig(); err != nil {
			return Config{}, fmt.Errorf("config: read yaml (%s): %w", configPath, err)
		}
	}

	envVpr := viper.NewWithOptions(viper.KeyDelimiter(delimiter))
	envVpr.SetConfigType("env")
	envVpr.SetConfigFile(envPath)
	if err := envVpr.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf("config: read env (%s): %w", envPath, err)
	}

	// .env file wins over YAML but loses to OS environment.
	for _, key := range envVpr.AllKeys() {
		envKey := strings.ToUpper(strings.ReplaceAll(key, ".", delimiter))
		if _, ok := os.LookupEnv(envKey); !ok {
			base.Set(key, envVpr.Get(key))
		}
	}

	base.AutomaticEnv()
	base.SetEnvKeyReplacer(strings.NewReplacer(".", delimiter))

	var cfg Config
	if err := base.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("config: unmarshal: %w", err)
	}
	if err := validate(cfg); err != nil {
		return Config{}, fmt.Errorf("config: %w", err)
	}
	return cfg, nil
}

func validate(cfg Config) error {
	if cfg.JWT.Secret == "" {
		return fmt.Errorf("JWT__SECRET is required")
	}
	if cfg.DB.DSN == "" {
		return fmt.Errorf("DB__DSN is required")
	}
	if cfg.HTTP.Port == 0 {
		return fmt.Errorf("HTTP__PORT is required")
	}
	if cfg.General.AppEnvironment == "production" && cfg.AuthorizationMock.Enabled {
		return fmt.Errorf("AUTHORIZATION_MOCK__ENABLED must not be true in production")
	}
	return nil
}
