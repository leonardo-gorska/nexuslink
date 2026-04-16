package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

// Config holds all configuration variables loaded from the environment
type Config struct {
	AppEnv         string `env:"APP_ENV" envDefault:"development"`
	HTTPPort       string `env:"HTTP_PORT" envDefault:"8080"`
	MetricsPort    string `env:"METRICS_PORT" envDefault:"9090"`
	DatabaseURL    string `env:"DATABASE_URL,required"`
	RedisURL       string `env:"REDIS_URL,required"`
	RabbitMQURL    string `env:"RABBITMQ_URL,required"`
	RateLimitRPS   int    `env:"RATE_LIMIT_RPS" envDefault:"100"`
	RateLimitBurst int    `env:"RATE_LIMIT_BURST" envDefault:"200"`
	HashLength     int    `env:"HASH_LENGTH" envDefault:"7"`
	LogLevel       string `env:"LOG_LEVEL" envDefault:"debug"`

	// Worker specific configurations
	BatchSize         int    `env:"BATCH_SIZE" envDefault:"500"`
	BatchTimeout      string `env:"BATCH_TIMEOUT" envDefault:"5s"`
	WorkerConcurrency int    `env:"WORKER_CONCURRENCY" envDefault:"4"`
}

// LoadFromEnv parses the OS environment variables into the Config struct.
// It will return an error if a required field is missing.
func LoadFromEnv() (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}
	return &cfg, nil
}
