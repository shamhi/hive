package main

import (
	"context"
	"fmt"
	"hive/pkg/logger"
	"hive/services/api/internal/app"
	"hive/services/api/internal/config"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
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
		fmt.Fprintf(os.Stderr, "failed to run app: %v\n", err)
	case <-quitCtx.Done():
		stopCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()

		a.Stop(stopCtx)
	}
}

func fatal(msg string, val any) {
	fmt.Fprintf(os.Stderr, "%s: %v\n", msg, val)
	os.Exit(1)
}
