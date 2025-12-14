package config

import (
	"flag"
	"hive/pkg/db/redis"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env      string `env:"ENV" env-default:"local"`
	GRPCPort int    `env:"GRPC_PORT" env-required:"true"`

	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" env-default:"5s"`

	SearchRadius float64 `env:"SEARCH_RADIUS" env-default:"30000"`

	RedisConfig redis.Config
}

var appConfig Config

func ParseFlags() (*Config, error) {
	flag.StringVar(&appConfig.Env, "env", appConfig.Env, "local | dev | prod")
	flag.IntVar(&appConfig.GRPCPort, "grpc_port", appConfig.GRPCPort, "gRPC port of this server")
	flag.DurationVar(&appConfig.ShutdownTimeout, "shutdown_timeout", appConfig.ShutdownTimeout, "time")
	flag.Float64Var(&appConfig.SearchRadius, "search_radius", appConfig.SearchRadius, "search radius in meters for store lookup")

	flag.StringVar(&appConfig.RedisConfig.Host, "redis_host", appConfig.RedisConfig.Host, "Redis host")
	flag.IntVar(&appConfig.RedisConfig.Port, "redis_port", appConfig.RedisConfig.Port, "Redis port")
	flag.StringVar(&appConfig.RedisConfig.Password, "redis_password", appConfig.RedisConfig.Password, "Redis password")
	flag.IntVar(&appConfig.RedisConfig.DB, "redis_db", appConfig.RedisConfig.DB, "Redis database number")

	if err := cleanenv.ReadEnv(&appConfig); err != nil {
		return nil, err
	}

	flag.Parse()

	return &appConfig, nil
}
