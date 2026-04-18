package app

import (
	"context"
	"fmt"
	trackingGen "hive/gen/tracking"
	"hive/pkg/grpcx"
	"hive/pkg/logger"
	"hive/services/tracking/internal/config"
	"hive/services/tracking/internal/service"
	"net"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type App struct {
	cfg        *config.ServerConfig
	lg         logger.Logger
	lis        net.Listener
	grpcServer *grpc.Server
	healthSrv  *health.Server
}

func New(
	cfg *config.ServerConfig,
	lg logger.Logger,
	repo service.DroneRepository,
) (*App, error) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %w", cfg.GRPCPort, err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpcx.UnaryServerLoggingTimeoutInterceptor(lg, 10*time.Second)),
	)

	trackingService := service.New(
		repo,
		lg,
	)
	trackingGen.RegisterTrackingServiceServer(grpcServer, trackingService)
	healthSrv := health.NewServer()
	healthSrv.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthSrv.SetServingStatus(trackingGen.TrackingService_ServiceDesc.ServiceName, healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(grpcServer, healthSrv)

	return &App{
		cfg:        cfg,
		lg:         lg,
		lis:        lis,
		grpcServer: grpcServer,
		healthSrv:  healthSrv,
	}, nil
}

func (a *App) Run(errChan chan<- error) {
	lg := a.lg.With(zap.String("component", "app"))

	lg.Info(context.Background(), "Running gRPC server",
		zap.Int("port", a.cfg.GRPCPort),
	)

	if err := a.grpcServer.Serve(a.lis); err != nil {
		errChan <- fmt.Errorf("failed to serve gRPC server: %w", err)
	}
}

func (a *App) Stop(ctx context.Context) {
	lg := a.lg.With(zap.String("component", "app"))
	lg.Info(ctx, "Gracefully shutting down...")
	if a.healthSrv != nil {
		a.healthSrv.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)
		a.healthSrv.SetServingStatus(trackingGen.TrackingService_ServiceDesc.ServiceName, healthpb.HealthCheckResponse_NOT_SERVING)
	}

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
}
