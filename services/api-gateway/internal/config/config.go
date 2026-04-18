package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env string `env:"ENV" env-default:"local"`

	ServerPort int `env:"SERVER_PORT" env-default:"8080"`

	RequestTimeout        time.Duration `env:"REQUEST_TIMEOUT" env-default:"15s"`
	ShutdownTimeout       time.Duration `env:"SHUTDOWN_TIMEOUT" env-default:"5s"`
	HTTPReadTimeout       time.Duration `env:"HTTP_READ_TIMEOUT" env-default:"15s"`
	HTTPReadHeaderTimeout time.Duration `env:"HTTP_READ_HEADER_TIMEOUT" env-default:"5s"`
	HTTPWriteTimeout      time.Duration `env:"HTTP_WRITE_TIMEOUT" env-default:"30s"`
	HTTPIdleTimeout       time.Duration `env:"HTTP_IDLE_TIMEOUT" env-default:"60s"`
	CORSAllowedOrigins    string        `env:"CORS_ALLOWED_ORIGINS" env-default:"http://localhost,http://127.0.0.1"`

	OrderAddr    string `env:"ORDER_ADDR" env-required:"true"`
	BaseAddr     string `env:"BASE_ADDR" env-required:"true"`
	StoreAddr    string `env:"STORE_ADDR" env-required:"true"`
	TrackingAddr string `env:"TRACKING_ADDR" env-required:"true"`
	DispatchAddr string `env:"DISPATCH_ADDR" env-required:"true"`
}

func Load() (*Config, error) {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("failed to read config from env: %w", err)
	}

	return &cfg, nil
}
