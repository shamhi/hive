package main

import (
	"context"
	"fmt"
	hredis "hive/pkg/db/redis"
	"hive/pkg/logger"
	"hive/services/tracking/internal/config"
	"hive/services/tracking/internal/repository"
	"hive/services/tracking/internal/transport/grpc"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx := context.Background()

	cfg, err := config.NewServerConfig()
	if err != nil {
		panic(fmt.Errorf("falied to load config: %w", err))
	}

	lg, err := logger.NewLogger(cfg.Env)
	if err != nil {
		panic(fmt.Errorf("falied to initialize logger: %w", err))
	}

	redisDb, err := hredis.New(cfg.RedisConfig)
	if err != nil {
		panic(fmt.Errorf("failed to connect to redis: %w", err))
	}

	server, err := grpc.New(&cfg.ServerConfig, &lg)
	if err != nil {
		panic(fmt.Errorf("failed to create server: %w", err))
	}

	repo := repository.New(&hredis.Database{Client: redisDb.Client})

	server.RegisterService(*repo)

	lg.Info(ctx, "starting grpc server")

	err = server.Start()
	if err != nil {
		panic(fmt.Errorf("failed to start server: %w", err))
	}
	lg.Info(ctx, fmt.Sprintf("GRPC server listening on %s:%d", cfg.ServerConfig.Host, cfg.ServerConfig.Port))

	graceSh := make(chan os.Signal, 1)
	signal.Notify(graceSh, os.Interrupt, syscall.SIGTERM)
	<-graceSh

	lg.Info(ctx, "Shutdown signal received, starting graceful shutdown...")
}
