package kafka

import (
	"context"
	"fmt"
	"hive/pkg/logger"
	"strings"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Consumer struct {
	Reader *kafka.Reader
	lg     logger.Logger
}

func NewConsumer(cfg Config, topic string, lg logger.Logger) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     strings.Split(cfg.Brokers, ","),
		GroupID:     cfg.GroupID,
		Topic:       topic,
		StartOffset: kafka.LastOffset,
	})

	return &Consumer{
		Reader: r,
		lg:     lg,
	}
}

func (c *Consumer) Start(ctx context.Context, handler func(context.Context, []byte) error) error {
	lg := c.lg.With(
		zap.String("component", "kafka-consumer"),
		zap.String("topic", c.Reader.Config().Topic),
	)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msg, err := c.Reader.ReadMessage(ctx)
			if err != nil {
				return fmt.Errorf("kafka consumer stopped: %w", err)
			}

			if err := handler(ctx, msg.Value); err != nil {
				lg.Error(ctx, "failed to handle message", zap.Error(err))
			}
		}
	}
}

func (c *Consumer) Close() error {
	return c.Reader.Close()
}
