package app

import (
	"context"
	"fmt"
	pbBase "hive/gen/base"
	"hive/pkg/grpcx"
	"hive/pkg/resilience"
	grpcStoreClient "hive/services/dispatch/internal/infrastructure/grpc/store"
	"time"

	pb "hive/gen/dispatch"
	pbOrder "hive/gen/order"
	pbStore "hive/gen/store"
	pbTelemetry "hive/gen/telemetry"
	pbTracking "hive/gen/tracking"
	pg "hive/pkg/db/postgres"
	"hive/pkg/kafka"
	"hive/pkg/logger"
	"hive/services/dispatch/internal/config"
	grpcBaseClient "hive/services/dispatch/internal/infrastructure/grpc/base"
	grpcOrderClient "hive/services/dispatch/internal/infrastructure/grpc/order"
	grpcTelemetryClient "hive/services/dispatch/internal/infrastructure/grpc/telemetry"
	grpcTrackingClient "hive/services/dispatch/internal/infrastructure/grpc/tracking"
	repoPostgres "hive/services/dispatch/internal/repository/postgres"
	"hive/services/dispatch/internal/service"
	transportGrpc "hive/services/dispatch/internal/transport/grpc"
	transportKafka "hive/services/dispatch/internal/transport/kafka"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type App struct {
	cfg            *config.Config
	lg             logger.Logger
	lis            net.Listener
	grpcServer     *grpc.Server
	grpcConns      []*grpc.ClientConn
	postgresDB     *pg.Database
	consumer       *kafka.Consumer
	handler        *transportKafka.Handler
	cancelConsumer context.CancelFunc
}

func New(cfg *config.Config, lg logger.Logger) (*App, error) {
	db, err := pg.New(cfg.DBConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	repo := repoPostgres.NewPostgresRepo(db.Pool)

	orderConn, err := grpc.NewClient(
		cfg.OrderAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(grpcx.UnaryClientResilienceInterceptor(lg, grpcx.ClientResilienceConfig{
			Name:    "dispatch->order",
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

	storeConn, err := grpc.NewClient(
		cfg.StoreAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(grpcx.UnaryClientResilienceInterceptor(lg, grpcx.ClientResilienceConfig{
			Name:    "dispatch->store",
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
		return nil, fmt.Errorf("failed to connect to store service: %w", err)
	}
	storeClient := grpcStoreClient.NewStoreClient(pbStore.NewStoreServiceClient(storeConn))

	baseConn, err := grpc.NewClient(
		cfg.BaseAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(grpcx.UnaryClientResilienceInterceptor(lg, grpcx.ClientResilienceConfig{
			Name:    "dispatch->base",
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
		storeConn.Close()
		return nil, fmt.Errorf("failed to connect to base service: %w", err)
	}
	baseClient := grpcBaseClient.NewBaseClient(pbBase.NewBaseServiceClient(baseConn))

	trackingConn, err := grpc.NewClient(
		cfg.TrackingAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(grpcx.UnaryClientResilienceInterceptor(lg, grpcx.ClientResilienceConfig{
			Name:    "dispatch->tracking",
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
		storeConn.Close()
		baseConn.Close()
		return nil, fmt.Errorf("failed to connect to tracking service: %w", err)
	}
	trackingClient := grpcTrackingClient.NewTrackingClient(pbTracking.NewTrackingServiceClient(trackingConn))

	telemetryConn, err := grpc.NewClient(
		cfg.TelemetryAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(grpcx.UnaryClientResilienceInterceptor(lg, grpcx.ClientResilienceConfig{
			Name:    "dispatch->tracking",
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
		storeConn.Close()
		baseConn.Close()
		trackingConn.Close()
		return nil, fmt.Errorf("failed to connect to telemetry service: %w", err)
	}
	telemetryClient := grpcTelemetryClient.NewTelemetryClient(pbTelemetry.NewTelemetryServiceClient(telemetryConn))

	dispatchService := service.NewDispatchService(
		repo,
		orderClient,
		storeClient,
		baseClient,
		trackingClient,
		telemetryClient,
		lg,
	)

	consumer := kafka.NewConsumer(cfg.KafkaConfig, cfg.TelemetryEventsTopic, lg)
	kafkaHandler := transportKafka.NewHandler(dispatchService)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		orderConn.Close()
		storeConn.Close()
		baseConn.Close()
		trackingConn.Close()
		telemetryConn.Close()
		return nil, fmt.Errorf("failed to listen on port %d: %w", cfg.GRPCPort, err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpcx.UnaryServerLoggingTimeoutInterceptor(lg, 10*time.Second)),
	)
	dispatchServer := transportGrpc.NewServer(
		dispatchService,
		&transportGrpc.Config{
			MinDroneBattery:   cfg.MinDroneBattery,
			DroneSearchRadius: cfg.DroneSearchRadius,
		},
	)
	pb.RegisterDispatchServiceServer(grpcServer, dispatchServer)

	return &App{
		cfg:        cfg,
		lg:         lg,
		lis:        lis,
		grpcServer: grpcServer,
		grpcConns:  []*grpc.ClientConn{orderConn, storeConn, baseConn, trackingConn, telemetryConn},
		postgresDB: db,
		consumer:   consumer,
		handler:    kafkaHandler,
	}, nil
}

func (a *App) Run(errChan chan<- error) {
	lg := a.lg.With(zap.String("component", "app"))

	consumerCtx, cancelConsumer := context.WithCancel(context.Background())
	a.cancelConsumer = cancelConsumer

	go a.runConsumer(consumerCtx)

	lg.Info(context.Background(), "Running gRPC server",
		zap.Int("port", a.cfg.GRPCPort),
		zap.String("env", a.cfg.Env),
	)

	if err := a.grpcServer.Serve(a.lis); err != nil {
		a.cancelConsumer()
		errChan <- fmt.Errorf("failed to serve gRPC server: %w", err)
	}
}

func (a *App) runConsumer(ctx context.Context) {
	lg := a.lg.With(
		zap.String("component", "kafka-consumer"),
	)

	if err := a.consumer.Start(ctx, a.handler.Handle); err != nil {
		lg.Error(ctx, "consumer exited with error", zap.Error(err))
	}

	lg.Info(ctx, "consumer stopped")
}

func (a *App) Stop(ctx context.Context) {
	lg := a.lg.With(zap.String("component", "app"))

	lg.Info(ctx, "Gracefully shutting down...")

	if a.cancelConsumer != nil {
		a.cancelConsumer()
		lg.Info(ctx, "Kafka consumer context cancelled")
	}

	if a.consumer != nil {
		a.consumer.Close()
	}
	lg.Info(ctx, "Kafka consumer closed")

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
