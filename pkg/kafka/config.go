package kafka

type Config struct {
	Brokers string `env:"KAFKA_BROKERS" env-required:"true"`
	GroupID string `env:"KAFKA_GROUP_ID" env-default:"default-group"`
}
