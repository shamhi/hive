package config

import (
	"flag"
	"hive/pkg/kafka"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env      string `env:"ENV" env-default:"local"`
	GRPCPort int    `env:"GRPC_PORT" env-required:"true"`

	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" env-default:"10s"`

	EventsTopic string `env:"EVENTS_TOPIC" env-default:"telemetry-events"`
	DataTopic   string `env:"DATA_TOPIC" env-default:"telemetry-data"`

	KafkaConfig kafka.Config
}

var appConfig Config

func ParseFlags() (*Config, error) {
	flag.StringVar(&appConfig.Env, "env", appConfig.Env, "local | dev | prod")
	flag.IntVar(&appConfig.GRPCPort, "grpc_port", appConfig.GRPCPort, "gRPC port of this server")
	flag.DurationVar(&appConfig.ShutdownTimeout, "shutdown_timeout", appConfig.ShutdownTimeout, "timeout of server shutdown")
	flag.StringVar(&appConfig.EventsTopic, "telemetry_topic", appConfig.EventsTopic, "kafka topic for telemetry events")
	flag.StringVar(&appConfig.DataTopic, "data_topic", appConfig.DataTopic, "kafka topic for telemetry data")

	flag.StringVar(&appConfig.KafkaConfig.Brokers, "kafka_brokers", appConfig.KafkaConfig.Brokers, "Kafka brokers, comma separated")
	flag.StringVar(&appConfig.KafkaConfig.GroupID, "kafka_group_id", appConfig.KafkaConfig.GroupID, "Kafka consumer group ID")

	if err := cleanenv.ReadEnv(&appConfig); err != nil {
		return nil, err
	}

	flag.Parse()

	return &appConfig, nil
}
