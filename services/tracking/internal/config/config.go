package config

import (
	"fmt"
	"hive/pkg/db/redis"
	"hive/pkg/kafka"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type ServerConfig struct {
	Env string `env:"ENV" env-default:"local"`

	GRPCPort int `env:"GRPC_PORT" env-default:"50055"`

	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" env-default:"5s"`

	RedisConfig redis.Config
}

type WorkerConfig struct {
	Env string `env:"ENV" env-default:"local"`

	DataTopic string `env:"TELEMETRY_DATA_TOPIC" env-default:"telemetry-data"`

	KafkaConfig kafka.Config
	RedisConfig redis.Config
}

func NewServerConfig() (*ServerConfig, error) {
	var cfg ServerConfig

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("failed to read server env: %w", err)
	}

	return &cfg, nil
}

func NewWorkerConfig() (*WorkerConfig, error) {
	var cfg WorkerConfig

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("failed to read server env: %w", err)
	}

	return &cfg, nil
}
