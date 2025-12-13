package config

import (
	"hive/pkg/db/redis"
	"hive/pkg/kafka"

	"github.com/caarlos0/env/v11"
)

type KafkaConfig struct {
	kafka.Config
	Topic string `env:"KAFKA_TOPIC,required"`
}

type GRPCConfig struct {
	Host string `env:"GRPCHOST,required"`
	Port int    `env:"GRPCPORT,required"`
}

type ServerConfig struct {
	Env          string `env:"ENV,required"`
	RedisConfig  redis.Config
	ServerConfig GRPCConfig
}

type WorkerConfig struct {
	Env         string `env:"ENV,required"`
	KafkaConfig KafkaConfig
	RedisConfig redis.Config
}

func NewServerConfig() (*ServerConfig, error) {
	var cfg ServerConfig
	err := env.Parse(&cfg)

	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func NewWorkerConfig() (*WorkerConfig, error) {
	var cfg WorkerConfig
	err := env.Parse(&cfg)

	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
