package config

import (
	"hive/pkg/db/redis"
	"hive/pkg/kafka"

	"github.com/ilyakaznacheev/cleanenv"
)

type KafkaAppConfig struct {
	kafka.Config
	Topic string `env:"KAFKA_TOPIC" env-required:"true"`
}

type GRPCServerConfig struct {
	Port int `env:"GRPC_PORT" env-default:"50055"`
}

type Config struct {
	Env          string `env:"ENV" env-default:"local"`
	KafkaConfig  KafkaAppConfig
	RedisConfig  redis.Config
	ServerConfig GRPCServerConfig
}

func New() (*Config, error) {
	var cfg Config
	err := cleanenv.ReadEnv(&cfg)

	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
