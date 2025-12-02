package config

import (
	"flag"
	"hive/pkg/db/postgres"
	"hive/pkg/kafka"
	"time"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	Env             string        `env:"ENV" env-default:"local"`
	GRPCPort        int           `env:"GRPC_PORT" env-required:"true"`
	OrderAddr       string        `env:"ORDER_ADDR" env-required:"true"`
	TrackingAddr    string        `env:"TRACKING_ADDR" env-required:"true"`
	TelemetryAddr   string        `env:"TELEMETRY_ADDR" env-required:"true"`
	RequestTimeout  time.Duration `env:"REQUEST_TIMEOUT" env-default:"30s"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" env-default:"10s"`
	KafkaConfig     kafka.Config
	DBConfig        postgres.Config
}

var appConfig Config

func ParseFlags() (*Config, error) {
	flag.StringVar(&appConfig.Env, "env", appConfig.Env, "local | dev | prod")
	flag.IntVar(&appConfig.GRPCPort, "grpc_port", appConfig.GRPCPort, "gRPC port of this server")
	flag.StringVar(&appConfig.OrderAddr, "order_addr", appConfig.OrderAddr, "address of order server")
	flag.StringVar(&appConfig.TrackingAddr, "tracking_addr", appConfig.TrackingAddr, "address of tracking server")
	flag.StringVar(&appConfig.TelemetryAddr, "telemetry_addr", appConfig.TelemetryAddr, "address of telemetry server")
	flag.DurationVar(&appConfig.RequestTimeout, "request_timeout", appConfig.RequestTimeout, "timeout of dispatch requesting")
	flag.DurationVar(&appConfig.ShutdownTimeout, "shutdown_timeout", appConfig.ShutdownTimeout, "time")
	flag.Parse()

	if err := env.Parse(&appConfig); err != nil {
		return nil, err
	}

	return &appConfig, nil
}
