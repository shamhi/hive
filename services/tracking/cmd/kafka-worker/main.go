package main

import (
	"context"
	"fmt"
	hredis "hive/pkg/db/redis"
	"hive/pkg/kafka"
	"hive/pkg/logger"
	"hive/services/tracking/internal/config"
	"hive/services/tracking/internal/repository/redis"
	tkafka "hive/services/tracking/internal/transport/kafka"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := config.NewWorkerConfig()
	if err != nil {
		fatal("failed to parse config", err)
	}

	lg, err := logger.NewLogger(cfg.Env)
	if err != nil {
		fatal("failed to initialize logger", err)
	}

	quitCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	redisDb, err := hredis.New(cfg.RedisConfig)
	if err != nil {
		fatal("failed to initialize redis client", err)
	}

	repo := redis.NewRedisRepo(redisDb)
	handler := tkafka.New(
		repo,
		lg,
	)

	kafkaConsumer := kafka.NewConsumer(cfg.KafkaConfig, cfg.DataTopic, lg)

	lg.Info(context.Background(), "starting kafka worker...")
	if err := kafkaConsumer.Start(quitCtx, handler.HandleMessage); err != nil {
		fatal("failed to start kafka consumer", err)
	}
	lg.Info(context.Background(), "kafka worker stopped gracefully")
}

func fatal(msg string, val any) {
	panic(fmt.Sprintf("%s: %v\n", msg, val))
}
