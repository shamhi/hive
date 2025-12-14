package main

import (
	"context"
	"fmt"
	hredis "hive/pkg/db/redis"
	"hive/pkg/logger"
	"hive/services/tracking/internal/config"
	"hive/services/tracking/internal/repository/redis"
	transportGrpc "hive/services/tracking/internal/transport/grpc"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := config.NewServerConfig()
	if err != nil {
		fatal("failed to parse config", err)
	}

	lg, err := logger.NewLogger(cfg.Env)
	if err != nil {
		fatal("failed to initialize logger", err)
	}

	redisDb, err := hredis.New(cfg.RedisConfig)
	if err != nil {
		fatal("failed to initialize redis client", err)
	}

	server, err := transportGrpc.New(cfg, lg, redis.NewRedisRepo(redisDb))
	if err != nil {
		fatal("failed to initialize server", err)
	}

	quitCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errChan := make(chan error, 1)
	go server.Run(errChan)

	select {
	case err := <-errChan:
		fatal("failed to run server", err)
	case <-quitCtx.Done():
		stopCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()

		server.Stop(stopCtx)
	}
}

func fatal(msg string, val any) {
	panic(fmt.Sprintf("%s: %v\n", msg, val))
}
