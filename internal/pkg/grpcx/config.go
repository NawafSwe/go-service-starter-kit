package grpcx

import "time"

type Config struct {
	Name           string               `mapstructure:"NAME"`
	Address        string               `mapstructure:"ADDRESS"`
	Timeout        time.Duration        `mapstructure:"TIMEOUT"`
	MaxRetries     int                  `mapstructure:"MAX_RETRIES"`
	RetryWaitMin   time.Duration        `mapstructure:"RETRY_WAIT_MIN"`
	RetryWaitMax   time.Duration        `mapstructure:"RETRY_WAIT_MAX"`
	CircuitBreaker CircuitBreakerConfig `mapstructure:"CIRCUIT_BREAKER"`
}

type CircuitBreakerConfig struct {
	MaxRequests uint32        `mapstructure:"MAX_REQUESTS"`
	Interval    time.Duration `mapstructure:"INTERVAL"`
	Timeout     time.Duration `mapstructure:"TIMEOUT"`
	Threshold   uint32        `mapstructure:"THRESHOLD"`
}
