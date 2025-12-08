package config

import (
	"flag"
	"hive/pkg/db/postgres"
	"hive/pkg/kafka"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Env      string `env:"ENV" env-default:"local"`
	GRPCPort int    `env:"GRPC_PORT" env-required:"true"`

	RequestTimeout  time.Duration `env:"REQUEST_TIMEOUT" env-default:"30s"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" env-default:"10s"`

	MinDroneBattery   float64 `env:"MIN_DRONE_BATTERY" env-default:"50"`
	DroneSearchRadius float64 `env:"DRONE_SEARCH_RADIUS" env-default:"30000"`

	TelemetryAddr string `env:"TELEMETRY_ADDR" env-required:"true"`
	OrderAddr     string `env:"ORDER_ADDR" env-required:"true"`
	StoreAddr     string `env:"STORE_ADDR" env-required:"true"`
	BaseAddr      string `env:"BASE_ADDR" env-required:"true"`
	TrackingAddr  string `env:"TRACKING_ADDR" env-required:"true"`

	TelemetryEventsTopic string `env:"TELEMETRY_EVENTS_TOPIC" env-default:"telemetry-events"`

	KafkaConfig kafka.Config
	DBConfig    postgres.Config
}

var appConfig Config

func ParseFlags() (*Config, error) {
	flag.StringVar(&appConfig.Env, "env", appConfig.Env, "local | dev | prod")
	flag.IntVar(&appConfig.GRPCPort, "grpc_port", appConfig.GRPCPort, "gRPC port of this server")
	flag.DurationVar(&appConfig.RequestTimeout, "request_timeout", appConfig.RequestTimeout, "timeout of dispatch requesting")
	flag.DurationVar(&appConfig.ShutdownTimeout, "shutdown_timeout", appConfig.ShutdownTimeout, "timeout of server shutdown")
	flag.Float64Var(&appConfig.MinDroneBattery, "min_battery", appConfig.MinDroneBattery, "minimum battery percentage for drone assignment")
	flag.Float64Var(&appConfig.DroneSearchRadius, "search_radius", appConfig.DroneSearchRadius, "search radius in meters for drone assignment")
	flag.StringVar(&appConfig.TelemetryAddr, "telemetry_addr", appConfig.TelemetryAddr, "address of telemetry server")
	flag.StringVar(&appConfig.OrderAddr, "order_addr", appConfig.OrderAddr, "address of order server")
	flag.StringVar(&appConfig.StoreAddr, "store_addr", appConfig.StoreAddr, "address of store server")
	flag.StringVar(&appConfig.BaseAddr, "base_addr", appConfig.BaseAddr, "address of base server")
	flag.StringVar(&appConfig.TrackingAddr, "tracking_addr", appConfig.TrackingAddr, "address of tracking server")
	flag.StringVar(&appConfig.TelemetryEventsTopic, "telemetry_events_topic", appConfig.TelemetryEventsTopic, "kafka topic for telemetry events")

	flag.StringVar(&appConfig.DBConfig.Host, "db_host", appConfig.DBConfig.Host, "Postgres host")
	flag.IntVar(&appConfig.DBConfig.Port, "db_port", appConfig.DBConfig.Port, "Postgres port")
	flag.StringVar(&appConfig.DBConfig.Username, "db_user", appConfig.DBConfig.Username, "Postgres user")
	flag.StringVar(&appConfig.DBConfig.Password, "db_password", appConfig.DBConfig.Password, "Postgres password")
	flag.StringVar(&appConfig.DBConfig.DBName, "db_name", appConfig.DBConfig.DBName, "Postgres database name")

	flag.StringVar(&appConfig.KafkaConfig.Brokers, "kafka_brokers", appConfig.KafkaConfig.Brokers, "Kafka brokers, comma separated")
	flag.StringVar(&appConfig.KafkaConfig.GroupID, "kafka_group_id", appConfig.KafkaConfig.GroupID, "Kafka consumer group ID")

	if err := env.Parse(&appConfig); err != nil {
		return nil, err
	}
	if err := env.Parse(&appConfig.DBConfig); err != nil {
		return nil, err
	}
	if err := env.Parse(&appConfig.KafkaConfig); err != nil {
		return nil, err
	}

	flag.Parse()

	return &appConfig, nil
}
