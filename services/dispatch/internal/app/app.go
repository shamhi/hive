package app

import (
	"context"
	"fmt"
	pb "hive/gen/dispatch"
	pbOrder "hive/gen/order"
	pbTelemetry "hive/gen/telemetry"
	pbTracking "hive/gen/tracking"
	pg "hive/pkg/db/postgres"
	"hive/pkg/logger"
	"hive/services/dispatch/internal/config"
	orderClient "hive/services/dispatch/internal/infrastructure/client/order"
	telemetryClient "hive/services/dispatch/internal/infrastructure/client/telemetry"
	trackingClient "hive/services/dispatch/internal/infrastructure/client/tracking"
	repoPostgres "hive/services/dispatch/internal/repository/postgres"
	"hive/services/dispatch/internal/service"
	transportGrpc "hive/services/dispatch/internal/transport/grpc"
	"net"
	"strconv"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type App struct {
	cfg        *config.Config
	lg         logger.Logger
	lis        net.Listener
	grpcServer *grpc.Server
}

func New(cfg *config.Config, lg logger.Logger) (*App, error) {
	db, err := pg.New(cfg.DBConfig)
	if err != nil {
		return nil, err
	}

	repo := repoPostgres.NewPostgresRepo(db.Pool)

	orderConn, err := grpc.NewClient(cfg.OrderAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to order service: %w", err)
	}
	orderAdapter := orderClient.NewOrderAdapter(pbOrder.NewOrderServiceClient(orderConn))

	trackingConn, err := grpc.NewClient(cfg.TrackingAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to tracking service: %w", err)
	}
	trackingAdapter := trackingClient.NewTrackingAdapter(pbTracking.NewTrackingServiceClient(trackingConn))

	telemetryConn, err := grpc.NewClient(cfg.TelemetryAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to telemetry service: %w", err)
	}
	telemetryAdapter := telemetryClient.NewTelemetryAdapter(pbTelemetry.NewTelemetryServiceClient(telemetryConn))

	dispatchService := service.NewDispatchService(
		repo,
		orderAdapter,
		trackingAdapter,
		telemetryAdapter,
	)

	lis, err := net.Listen("tcp", ":"+strconv.Itoa(cfg.GRPCPort))
	if err != nil {
		return nil, err
	}

	grpcServer := grpc.NewServer()
	dispatchServer := transportGrpc.NewServer(dispatchService)
	pb.RegisterDispatchServiceServer(grpcServer, dispatchServer)

	return &App{
		cfg:        cfg,
		lg:         lg,
		lis:        lis,
		grpcServer: grpcServer,
	}, nil
}

func (a *App) Run(errChan chan<- error) {
	lg := a.lg.With(zap.String("op", "app.Run"))

	lg.Info(context.Background(), "Running gRPC server", zap.Int("port", a.cfg.GRPCPort))
	if err := a.grpcServer.Serve(a.lis); err != nil {
		errChan <- err
	}
}

func (a *App) Stop(ctx context.Context) {
	lg := a.lg.With(zap.String("op", "app.Stop"))

	lg.Info(context.Background(), "Shutting down...")

	done := make(chan struct{})
	go func() {
		a.grpcServer.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		lg.Info(context.Background(), "gRPC server gracefully stopped")
	case <-ctx.Done():
		lg.Warn(context.Background(), "Timeout reached, force stopping...")
		a.grpcServer.Stop()
		lg.Info(context.Background(), "gRPC server force stopped")
	}
}
