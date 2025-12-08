package config

import (
	"flag"
	"hive/pkg/db/postgres"
	"time"

	"github.com/caarlos0/env/v11"
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
	flag.DurationVar(&appConfig.RequestTimeout, "request_timeout", appConfig.RequestTimeout, "timeout of dispatch requesting")
	flag.DurationVar(&appConfig.ShutdownTimeout, "shutdown_timeout", appConfig.ShutdownTimeout, "time")
	flag.StringVar(&appConfig.DispatchAddr, "dispatch_addr", appConfig.DispatchAddr, "address of dispatch server")

	flag.StringVar(&appConfig.DBConfig.Host, "db_host", appConfig.DBConfig.Host, "Postgres host")
	flag.IntVar(&appConfig.DBConfig.Port, "db_port", appConfig.DBConfig.Port, "Postgres port")
	flag.StringVar(&appConfig.DBConfig.Username, "db_user", appConfig.DBConfig.Username, "Postgres user")
	flag.StringVar(&appConfig.DBConfig.Password, "db_password", appConfig.DBConfig.Password, "Postgres password")
	flag.StringVar(&appConfig.DBConfig.DBName, "db_name", appConfig.DBConfig.DBName, "Postgres database name")

	if err := env.Parse(&appConfig); err != nil {
		return nil, err
	}
	if err := env.Parse(&appConfig.DBConfig); err != nil {
		return nil, err
	}

	flag.Parse()

	return &appConfig, nil
}
