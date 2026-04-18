package app

import (
	"context"
	"errors"
	"fmt"
	pbBase "hive/gen/base"
	pbDispatch "hive/gen/dispatch"
	pbOrder "hive/gen/order"
	pbStore "hive/gen/store"
	pbTracking "hive/gen/tracking"
	"hive/pkg/grpcx"
	"hive/pkg/logger"
	"hive/pkg/resilience"
	"hive/services/api-gateway/internal/config"
	grpcBaseClient "hive/services/api-gateway/internal/infrastructure/grpc/base"
	grpcDispatchClient "hive/services/api-gateway/internal/infrastructure/grpc/dispatch"
	grpcOrderClient "hive/services/api-gateway/internal/infrastructure/grpc/order"
	grpcStoreClient "hive/services/api-gateway/internal/infrastructure/grpc/store"
	grpcTrackingClient "hive/services/api-gateway/internal/infrastructure/grpc/tracking"
	apiV1 "hive/services/api-gateway/internal/transport/http/v1"
	"net/http"
	"strings"
	"time"

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
	allowedOrigins := parseAllowedOrigins(cfg.CORSAllowedOrigins)
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"http://localhost", "http://127.0.0.1"}
	}

	e.Validator = apiV1.NewCustomValidator()
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${time_rfc3339} | ${method} | ${uri} | ${status} | ${latency_human}\n",
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: allowedOrigins,
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.PATCH, echo.DELETE, echo.OPTIONS},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
			echo.HeaderXRequestedWith,
			echo.HeaderAccessControlRequestMethod,
			echo.HeaderAccessControlRequestHeaders,
		},
		ExposeHeaders: []string{echo.HeaderXRequestID},
		MaxAge:        86400,
	}))

	orderConn, err := grpc.NewClient(
		cfg.OrderAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(grpcx.UnaryClientResilienceInterceptor(lg, grpcx.ClientResilienceConfig{
			Name:    "api->order",
			Timeout: cfg.RequestTimeout,
			Retry: resilience.RetryConfig{
				MaxAttempts: 3,
				BaseDelay:   50 * time.Millisecond,
				MaxDelay:    500 * time.Millisecond,
				Jitter:      0.2,
			},
			Breaker: resilience.BreakerConfig{
				Interval:    10 * time.Second,
				Timeout:     5 * time.Second,
				MaxRequests: 3,
				MinRequests: 5,
				FailureRate: 0.6,
			},
		})),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to order service: %w", err)
	}
	orderClient := grpcOrderClient.NewOrderClient(pbOrder.NewOrderServiceClient(orderConn))

	baseConn, err := grpc.NewClient(
		cfg.BaseAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(grpcx.UnaryClientResilienceInterceptor(lg, grpcx.ClientResilienceConfig{
			Name:    "api->base",
			Timeout: cfg.RequestTimeout,
			Retry: resilience.RetryConfig{
				MaxAttempts: 3,
				BaseDelay:   50 * time.Millisecond,
				MaxDelay:    500 * time.Millisecond,
				Jitter:      0.2,
			},
			Breaker: resilience.BreakerConfig{
				Interval:    10 * time.Second,
				Timeout:     5 * time.Second,
				MaxRequests: 3,
				MinRequests: 5,
				FailureRate: 0.6,
			},
		})),
	)
	if err != nil {
		orderConn.Close()
		return nil, fmt.Errorf("failed to connect to base service: %w", err)
	}
	baseClient := grpcBaseClient.NewBaseClient(pbBase.NewBaseServiceClient(baseConn))

	storeConn, err := grpc.NewClient(
		cfg.StoreAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(grpcx.UnaryClientResilienceInterceptor(lg, grpcx.ClientResilienceConfig{
			Name:    "api->store",
			Timeout: cfg.RequestTimeout,
			Retry: resilience.RetryConfig{
				MaxAttempts: 3,
				BaseDelay:   50 * time.Millisecond,
				MaxDelay:    500 * time.Millisecond,
				Jitter:      0.2,
			},
			Breaker: resilience.BreakerConfig{
				Interval:    10 * time.Second,
				Timeout:     5 * time.Second,
				MaxRequests: 3,
				MinRequests: 5,
				FailureRate: 0.6,
			},
		})),
	)
	if err != nil {
		orderConn.Close()
		baseConn.Close()
		return nil, fmt.Errorf("failed to connect to store service: %w", err)
	}
	storeClient := grpcStoreClient.NewStoreClient(pbStore.NewStoreServiceClient(storeConn))

	trackingConn, err := grpc.NewClient(
		cfg.TrackingAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(grpcx.UnaryClientResilienceInterceptor(lg, grpcx.ClientResilienceConfig{
			Name:    "api->tracking",
			Timeout: cfg.RequestTimeout,
			Retry: resilience.RetryConfig{
				MaxAttempts: 3,
				BaseDelay:   50 * time.Millisecond,
				MaxDelay:    500 * time.Millisecond,
				Jitter:      0.2,
			},
			Breaker: resilience.BreakerConfig{
				Interval:    10 * time.Second,
				Timeout:     5 * time.Second,
				MaxRequests: 3,
				MinRequests: 5,
				FailureRate: 0.6,
			},
		})),
	)
	if err != nil {
		orderConn.Close()
		baseConn.Close()
		storeConn.Close()
		return nil, fmt.Errorf("failed to connect to tracking service: %w", err)
	}
	trackingClient := grpcTrackingClient.NewTrackingClient(pbTracking.NewTrackingServiceClient(trackingConn))

	dispatchConn, err := grpc.NewClient(
		cfg.DispatchAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(grpcx.UnaryClientResilienceInterceptor(lg, grpcx.ClientResilienceConfig{
			Name:    "api->dispatch",
			Timeout: cfg.RequestTimeout,
			Retry: resilience.RetryConfig{
				MaxAttempts: 3,
				BaseDelay:   50 * time.Millisecond,
				MaxDelay:    500 * time.Millisecond,
				Jitter:      0.2,
			},
			Breaker: resilience.BreakerConfig{
				Interval:    10 * time.Second,
				Timeout:     5 * time.Second,
				MaxRequests: 3,
				MinRequests: 5,
				FailureRate: 0.6,
			},
		})),
	)
	if err != nil {
		orderConn.Close()
		baseConn.Close()
		storeConn.Close()
		trackingConn.Close()
		return nil, fmt.Errorf("failed to connect to dispatch service: %w", err)
	}
	dispatchClient := grpcDispatchClient.NewDispatchClient(pbDispatch.NewDispatchServiceClient(dispatchConn))

	handler := apiV1.NewHandler(
		orderClient,
		baseClient,
		storeClient,
		trackingClient,
		dispatchClient,
	)
	apiV1.RegisterRoutes(e, handler)
	e.Server.ReadTimeout = cfg.HTTPReadTimeout
	e.Server.ReadHeaderTimeout = cfg.HTTPReadHeaderTimeout
	e.Server.WriteTimeout = cfg.HTTPWriteTimeout
	e.Server.IdleTimeout = cfg.HTTPIdleTimeout

	return &App{
		cfg:       cfg,
		lg:        lg,
		e:         e,
		grpcConns: []*grpc.ClientConn{orderConn, baseConn, storeConn, trackingConn, dispatchConn},
	}, nil
}

func (a *App) Run(errChan chan<- error) {
	lg := a.lg.With(zap.String("component", "app"))

	addr := fmt.Sprintf(":%d", a.cfg.ServerPort)
	lg.Info(context.Background(), "Running HTTP server",
		zap.String("addr", addr),
		zap.String("env", a.cfg.Env),
	)

	if err := a.e.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
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

func parseAllowedOrigins(raw string) []string {
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		origin := strings.TrimSpace(part)
		if origin == "" {
			continue
		}
		out = append(out, origin)
	}

	return out
}
