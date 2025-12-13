package main

import (
	"context"
	"fmt"
	hredis "hive/pkg/db/redis"
	"hive/pkg/kafka"
	"hive/pkg/logger"
	"hive/services/tracking/internal/config"
	"hive/services/tracking/internal/repository"
	tkafka "hive/services/tracking/internal/transport/kafka"
)

func main() {
	cfg, err := config.NewWorkerConfig()
	if err != nil {
		panic(fmt.Errorf("falied to load config: %w", err))
	}

	lg, err := logger.NewLogger(cfg.Env)
	if err != nil {
		panic(fmt.Errorf("falied to initialize logger: %w", err))
	}

	ctx := context.Background()

	redisDb, err := hredis.New(cfg.RedisConfig)
	if err != nil {
		panic(fmt.Errorf("failed to connect to redis: %w", err))
	}

	trackingRepo := repository.New(redisDb)

	handler := tkafka.New(trackingRepo)

	kafkaConsumer := kafka.NewConsumer(cfg.KafkaConfig.Config, cfg.KafkaConfig.Topic, lg)
	if err := kafkaConsumer.Start(ctx, handler.HandleMessage); err != nil {
		panic(fmt.Errorf("failed to start kafka consumer: %w", err))
	}

	kafkaConsumer.Close()
	redisDb.Close()
}
