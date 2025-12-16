package main

import (
	"context"
	"fmt"
	"hive/pkg/logger"
	"hive/services/api-gateway/internal/app"
	"hive/services/api-gateway/internal/config"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fatal("failed to load config", err)
	}

	lg, err := logger.NewLogger(cfg.Env)
	if err != nil {
		fatal("failed to create logger: %v\n", err)
	}

	a, err := app.New(cfg, lg)
	if err != nil {
		fatal("failed to create app: %v\n", err)
	}

	quitCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errChan := make(chan error, 1)
	go a.Run(errChan)

	select {
	case err := <-errChan:
		fatal("failed to run app", err)
	case <-quitCtx.Done():
		stopCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()

		a.Stop(stopCtx)
	}
}

func fatal(msg string, val any) {
	panic(fmt.Sprintf("%s: %v\n", msg, val))
}
