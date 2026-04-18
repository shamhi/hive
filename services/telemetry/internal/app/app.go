package app

import (
	"context"
	"fmt"
	pb "hive/gen/telemetry"
	"hive/pkg/grpcx"
	"hive/pkg/kafka"
	"hive/pkg/logger"
	"hive/services/telemetry/internal/config"
	"hive/services/telemetry/internal/service"
	transportGrpc "hive/services/telemetry/internal/transport/grpc"
	"net"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
)

type App struct {
	cfg            *config.Config
	lg             logger.Logger
	lis            net.Listener
	grpcServer     *grpc.Server
	healthSrv      *health.Server
	eventsProducer *kafka.Producer
	dataProducer   *kafka.Producer
}

func New(cfg *config.Config, lg logger.Logger) (*App, error) {
	eventsProducer := kafka.NewProducer(cfg.KafkaConfig)
	dataProducer := kafka.NewProducer(cfg.KafkaConfig)

	telemetryService := service.NewTelemetryService(
		eventsProducer.Writer,
		dataProducer.Writer,
		cfg.EventsTopic,
		cfg.DataTopic,
		lg,
	)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		eventsProducer.Close()
		dataProducer.Close()
		return nil, fmt.Errorf("failed to listen on port %d: %w", cfg.GRPCPort, err)
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(grpcx.UnaryServerLoggingTimeoutInterceptor(lg, 10*time.Second)),
		grpc.ChainStreamInterceptor(grpcx.LoggingTimeoutStreamServerInterceptor(lg, 0)),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    30 * time.Second,
			Timeout: 10 * time.Second,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             10 * time.Second,
			PermitWithoutStream: true,
		}),
	)
	telemetryServer := transportGrpc.NewServer(telemetryService)
	pb.RegisterTelemetryServiceServer(grpcServer, telemetryServer)
	healthSrv := health.NewServer()
	healthSrv.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthSrv.SetServingStatus(pb.TelemetryService_ServiceDesc.ServiceName, healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(grpcServer, healthSrv)

	return &App{
		cfg:            cfg,
		lg:             lg,
		lis:            lis,
		grpcServer:     grpcServer,
		healthSrv:      healthSrv,
		eventsProducer: eventsProducer,
		dataProducer:   dataProducer,
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
	if a.healthSrv != nil {
		a.healthSrv.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)
		a.healthSrv.SetServingStatus(pb.TelemetryService_ServiceDesc.ServiceName, healthpb.HealthCheckResponse_NOT_SERVING)
	}

	if a.eventsProducer != nil {
		a.eventsProducer.Close()
	}
	if a.dataProducer != nil {
		a.dataProducer.Close()
	}
	lg.Info(ctx, "Kafka producers closed")

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
