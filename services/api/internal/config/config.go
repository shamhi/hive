package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env string `env:"ENV" env-default:"local"`

	ServerPort int `env:"SERVER_PORT" env-default:"8080"`

	RequestTimeout  time.Duration `env:"REQUEST_TIMEOUT" env-default:"15s"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" env-default:"5s"`

	OrderAddr string `env:"ORDER_ADDR" env-required:"true"`
}

func Load() (*Config, error) {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("failed to read config from env: %w", err)
	}

	return &cfg, nil
}
