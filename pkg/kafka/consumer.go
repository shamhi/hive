package kafka

import (
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	Reader *kafka.Reader
}

func NewConsumer(cfg Config, topic string) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: cfg.Brokers,
		GroupID: cfg.GroupID,
		Topic:   topic,
	})

	return &Consumer{Reader: r}
}

func (c *Consumer) Close() error {
	return c.Reader.Close()
}
