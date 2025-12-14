package app

import (
	"context"
	"errors"
	"fmt"
	pbOrder "hive/gen/order"
	"hive/pkg/logger"
	"hive/services/api/internal/config"
	grpcOrderClient "hive/services/api/internal/infrastructure/client/order"
	"hive/services/api/internal/interceptor"
	apiV1 "hive/services/api/internal/transport/http/v1"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type App struct {
	cfg       *config.Config
	lg        logger.Logger
	e         *echo.Echo
	grpcConns []*grpc.ClientConn
}

func New(cfg *config.Config, lg logger.Logger) (*App, error) {
	e := echo.New()

	e.Validator = apiV1.NewCustomValidator()

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${time_rfc3339} | ${method} | ${uri} | ${status} | ${latency_human}\n",
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.PATCH, echo.DELETE, echo.OPTIONS},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	orderConn, err := grpc.NewClient(
		cfg.OrderAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(interceptor.TimeoutUnaryClientInterceptor(lg, cfg.RequestTimeout)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to order service: %w", err)
	}
	orderClient := grpcOrderClient.NewOrderClient(pbOrder.NewOrderServiceClient(orderConn))

	handler := apiV1.NewHandler(orderClient)
	apiV1.RegisterRoutes(e, handler)

	return &App{
		cfg:       cfg,
		lg:        lg,
		e:         e,
		grpcConns: []*grpc.ClientConn{orderConn},
	}, nil
}

func (a *App) Run(errChan chan<- error) {
	lg := a.lg.With(zap.String("component", "app"))

	addr := fmt.Sprintf(":%d", a.cfg.ServerPort)
	lg.Info(context.Background(), "Running HTTP server",
		zap.String("addr", addr),
		zap.String("env", a.cfg.Env),
	)

	if err := a.e.Start(addr); err != nil && errors.Is(err, http.ErrServerClosed) {
		errChan <- fmt.Errorf("failed to serve HTTP server: %w", err)
	}
}

func (a *App) Stop(ctx context.Context) {
	lg := a.lg.With(zap.String("component", "app"))

	lg.Info(ctx, "Gracefully shutting down...")

	if err := a.e.Shutdown(ctx); err != nil {
		lg.Warn(ctx, "failed to shutdown HTTP server", zap.Error(err))
	} else {
		lg.Info(ctx, "HTTP server gracefully stopped")
	}

	for _, conn := range a.grpcConns {
		if err := conn.Close(); err != nil {
			lg.Warn(ctx, "failed to close gRPC connection", zap.Error(err))
		}
	}
	lg.Info(ctx, "gRPC connections closed")
}
