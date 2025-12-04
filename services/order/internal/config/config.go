package config

import (
	"flag"
	"hive/pkg/db/postgres"
	"time"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	Env      string `env:"ENV" env-default:"local"`
	GRPCPort int    `env:"GRPC_PORT" env-required:"true"`

	RequestTimeout  time.Duration `env:"REQUEST_TIMEOUT" env-default:"30s"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" env-default:"10s"`

	DispatchAddr string `env:"DISPATCH_ADDR" env-required:"true"`

	DBConfig postgres.Config
}

var appConfig Config

func ParseFlags() (*Config, error) {
	flag.StringVar(&appConfig.Env, "env", appConfig.Env, "local | dev | prod")
	flag.IntVar(&appConfig.GRPCPort, "grpc_port", appConfig.GRPCPort, "gRPC port of this server")
	flag.StringVar(&appConfig.DispatchAddr, "dispatch_addr", appConfig.DispatchAddr, "address of dispatch server")
	flag.DurationVar(&appConfig.RequestTimeout, "request_timeout", appConfig.RequestTimeout, "timeout of dispatch requesting")
	flag.DurationVar(&appConfig.ShutdownTimeout, "shutdown_timeout", appConfig.ShutdownTimeout, "time")
	flag.Parse()

	if err := env.Parse(&appConfig); err != nil {
		return nil, err
	}

	return &appConfig, nil
}
