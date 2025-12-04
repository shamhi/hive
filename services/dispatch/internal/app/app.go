package app

import (
	"context"
	"fmt"

	pb "hive/gen/dispatch"
	pbOrder "hive/gen/order"
	pbStore "hive/gen/store"
	pbTelemetry "hive/gen/telemetry"
	pbTracking "hive/gen/tracking"
	pg "hive/pkg/db/postgres"
	"hive/pkg/kafka"
	"hive/pkg/logger"
	"hive/services/dispatch/internal/config"
	grpcOrderClient "hive/services/dispatch/internal/infrastructure/grpc/order"
	grpcStoreClient "hive/services/dispatch/internal/infrastructure/grpc/store"
	grpcTelemetryClient "hive/services/dispatch/internal/infrastructure/grpc/telemetry"
	grpcTrackingClient "hive/services/dispatch/internal/infrastructure/grpc/tracking"
	repoPostgres "hive/services/dispatch/internal/repository/postgres"
	"hive/services/dispatch/internal/service"
	transportGrpc "hive/services/dispatch/internal/transport/grpc"
	transportKafka "hive/services/dispatch/internal/transport/kafka"
	"net"
	"strconv"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type App struct {
	cfg            *config.Config
	lg             logger.Logger
	lis            net.Listener
	grpcServer     *grpc.Server
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
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to order service: %w", err)
	}
	orderClient := grpcOrderClient.NewOrderClient(pbOrder.NewOrderServiceClient(orderConn))

	storeConn, err := grpc.NewClient(
		cfg.StoreAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to store service: %w", err)
	}
	storeClient := grpcStoreClient.NewStoreClient(pbStore.NewStoreServiceClient(storeConn))

	trackingConn, err := grpc.NewClient(
		cfg.TrackingAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to tracking service: %w", err)
	}
	trackingClient := grpcTrackingClient.NewTrackingClient(pbTracking.NewTrackingServiceClient(trackingConn))

	telemetryConn, err := grpc.NewClient(
		cfg.TelemetryAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to telemetry service: %w", err)
	}
	telemetryClient := grpcTelemetryClient.NewTelemetryClient(pbTelemetry.NewTelemetryServiceClient(telemetryConn))

	dispatchService := service.NewDispatchService(
		repo,
		orderClient,
		storeClient,
		trackingClient,
		telemetryClient,
	)

	consumer := kafka.NewConsumer(cfg.KafkaConfig, cfg.TelemetryTopic, lg)
	kafkaHandler := transportKafka.NewHandler(dispatchService)

	lis, err := net.Listen("tcp", ":"+strconv.Itoa(cfg.GRPCPort))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %w", cfg.GRPCPort, err)
	}

	grpcServer := grpc.NewServer()
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
		consumer:   consumer,
		handler:    kafkaHandler,
	}, nil
}

func (a *App) Run(errChan chan<- error) {
	lg := a.lg.With(zap.String("component", "app"))

	consumerCtx, cancelConsumer := context.WithCancel(context.Background())
	a.cancelConsumer = cancelConsumer

	go func(ctx context.Context, lg logger.Logger) {
		lg = lg.With(zap.String("component", "kafka-consumer"))

		if err := a.consumer.Start(ctx, a.handler.Handle); err != nil {
			lg.Error(ctx, "consumer exited with error", zap.Error(err))
		}

		lg.Info(ctx, "consumer stopped")
	}(consumerCtx, lg)

	lg.Info(context.Background(), "Running gRPC server",
		zap.Int("port", a.cfg.GRPCPort),
		zap.String("env", a.cfg.Env),
	)

	if err := a.grpcServer.Serve(a.lis); err != nil {
		errChan <- fmt.Errorf("failed to serve gRPC server: %w", err)
	}
}

func (a *App) Stop(ctx context.Context) {
	lg := a.lg.With(zap.String("component", "app"), zap.String("operation", "shutdown"))

	lg.Info(ctx, "Gracefully shutting down...")

	if a.cancelConsumer != nil {
		a.cancelConsumer()
	}

	a.consumer.Close()
	lg.Info(ctx, "Kafka consumer closed")

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
