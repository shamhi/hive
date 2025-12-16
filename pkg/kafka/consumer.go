package kafka

import (
	"context"
	"fmt"
	"hive/pkg/logger"
	"hive/pkg/resilience"
	"strings"
	"time"

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
		zap.String("component", "kafka_consumer"),
		zap.String("topic", c.Reader.Config().Topic),
	)

	retryCfg := resilience.RetryConfig{
		MaxAttempts: 5,
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    2 * time.Second,
		Jitter:      0.2,
	}

	shouldRetry := func(err error) bool {
		if err == nil {
			return false
		}
		if ctx.Err() != nil {
			return false
		}
		return true
	}

	lg.Info(ctx, "consumer started")

	for {
		fetchStart := time.Now()
		msg, err := c.Reader.FetchMessage(ctx)
		if err != nil {
			lg.Error(ctx, "fetch message failed / consumer stopped",
				zap.Error(err),
				zap.Duration("duration", time.Since(fetchStart)),
			)
			return fmt.Errorf("kafka consumer stopped: %w", err)
		}

		lgMsg := lg.With(
			zap.Int("partition", msg.Partition),
			zap.Int64("offset", msg.Offset),
			zap.Int("key_len", len(msg.Key)),
			zap.Int("value_len", len(msg.Value)),
			zap.Time("time", msg.Time),
		)

		lgMsg.Info(ctx, "message fetched", zap.Duration("fetch_duration", time.Since(fetchStart)))

		handleStart := time.Now()
		if err := handler(ctx, msg.Value); err != nil {
			lgMsg.Error(ctx, "handler failed (message will not be committed)",
				zap.Error(err),
				zap.Duration("handle_duration", time.Since(handleStart)),
			)
			continue
		}
		lgMsg.Info(ctx, "handler completed", zap.Duration("handle_duration", time.Since(handleStart)))

		commitStart := time.Now()
		commitAttempts := 0
		err = resilience.Retry(ctx, retryCfg, shouldRetry, func(ctx context.Context) error {
			commitAttempts++
			return c.Reader.CommitMessages(ctx, msg)
		})
		if err != nil {
			lgMsg.Error(ctx, "commit failed after retries (consumer will stop)",
				zap.Int("attempts", commitAttempts),
				zap.Duration("commit_duration", time.Since(commitStart)),
				zap.Error(err),
			)
			return fmt.Errorf("failed to commit message: %w", err)
		}

		lgMsg.Info(ctx, "message committed",
			zap.Int("attempts", commitAttempts),
			zap.Duration("commit_duration", time.Since(commitStart)),
		)
	}
}

func (c *Consumer) Close() error {
	return c.Reader.Close()
}
