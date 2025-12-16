package app

import (
	"context"
	"fmt"
	pbDispatch "hive/gen/dispatch"
	pbOrder "hive/gen/order"
	pg "hive/pkg/db/postgres"
	"hive/pkg/grpcx"
	"hive/pkg/logger"
	"hive/pkg/resilience"
	"hive/services/order/internal/config"
	grpcDispatchClient "hive/services/order/internal/infrastructure/grpc/dispatch"
	repoPostgres "hive/services/order/internal/repository/postgres"
	"hive/services/order/internal/service"
	transportGrpc "hive/services/order/internal/transport/grpc"
	"net"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type App struct {
	cfg        *config.Config
	lg         logger.Logger
	lis        net.Listener
	grpcServer *grpc.Server
	grpcConns  []*grpc.ClientConn
	postgresDB *pg.Database
}

func New(cfg *config.Config, lg logger.Logger) (*App, error) {
	db, err := pg.New(cfg.DBConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}

	repo := repoPostgres.NewPostgresRepo(db.Pool)

	dispatchConn, err := grpc.NewClient(
		cfg.DispatchAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(grpcx.UnaryClientResilienceInterceptor(lg, grpcx.ClientResilienceConfig{
			Name:    "order->dispatch",
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
		return nil, fmt.Errorf("failed to connect to dispatch service: %w", err)
	}
	dispatchClient := grpcDispatchClient.NewDispatchClient(pbDispatch.NewDispatchServiceClient(dispatchConn))

	orderService := service.NewOrderService(
		repo,
		dispatchClient,
		lg,
	)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %w", cfg.GRPCPort, err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpcx.UnaryServerLoggingTimeoutInterceptor(lg, 10*time.Second)),
	)
	orderServer := transportGrpc.NewServer(orderService)
	pbOrder.RegisterOrderServiceServer(grpcServer, orderServer)

	return &App{
		cfg:        cfg,
		lg:         lg,
		lis:        lis,
		grpcServer: grpcServer,
		grpcConns:  []*grpc.ClientConn{dispatchConn},
		postgresDB: db,
	}, nil
}

func (a *App) Run(errChan chan<- error) {
	lg := a.lg.With(zap.String("component", "app"))

	lg.Info(context.Background(), "Running gRPC server",
		zap.Int("port", a.cfg.GRPCPort),
		zap.String("env", a.cfg.Env),
	)

	if err := a.grpcServer.Serve(a.lis); err != nil {
		errChan <- fmt.Errorf("failed to serve gRPC server: %w", err)
	}
}

func (a *App) Stop(ctx context.Context) {
	lg := a.lg.With(zap.String("component", "app"))

	lg.Info(ctx, "Gracefully shutting down...")

	for _, conn := range a.grpcConns {
		if err := conn.Close(); err != nil {
			lg.Warn(ctx, "failed to close gRPC connection", zap.Error(err))
		}
	}
	lg.Info(ctx, "gRPC connections closed")

	if a.postgresDB != nil {
		a.postgresDB.Close()
	}
	lg.Info(ctx, "Postgres database connection closed")

	done := make(chan struct{})
	go func() {
		a.grpcServer.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		lg.Info(ctx, "gRPC server gracefully stopped")
	case <-ctx.Done():
		lg.Warn(ctx, "shutdown timeout reached, force stopping gRPC server")
		a.grpcServer.Stop()
		lg.Info(ctx, "gRPC server force stopped")
	}

	lg.Info(ctx, "Shutdown completed successfully")
}
