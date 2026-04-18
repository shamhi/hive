package app

import (
	"context"
	"fmt"
	pb "hive/gen/base"
	"hive/pkg/db/redis"
	"hive/pkg/grpcx"
	"hive/pkg/logger"
	"hive/services/base/internal/config"
	repoRedis "hive/services/base/internal/repository/redis"
	"hive/services/base/internal/service"
	transportGrpc "hive/services/base/internal/transport/grpc"
	"net"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type App struct {
	cfg        *config.Config
	lg         logger.Logger
	lis        net.Listener
	grpcServer *grpc.Server
	healthSrv  *health.Server
	redisDB    *redis.Database
}

func New(cfg *config.Config, lg logger.Logger) (*App, error) {
	db, err := redis.New(cfg.RedisConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	repo := repoRedis.NewRedisRepo(db.Client)

	baseService := service.NewBaseService(
		repo,
		lg,
	)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %w", cfg.GRPCPort, err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpcx.UnaryServerLoggingTimeoutInterceptor(lg, 10*time.Second)),
	)
	baseServer := transportGrpc.NewServer(
		baseService,
		&transportGrpc.Config{SearchRadius: cfg.SearchRadius},
	)
	pb.RegisterBaseServiceServer(grpcServer, baseServer)
	healthSrv := health.NewServer()
	healthSrv.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthSrv.SetServingStatus(pb.BaseService_ServiceDesc.ServiceName, healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(grpcServer, healthSrv)

	return &App{
		cfg:        cfg,
		lg:         lg,
		lis:        lis,
		grpcServer: grpcServer,
		healthSrv:  healthSrv,
		redisDB:    db,
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
		a.healthSrv.SetServingStatus(pb.BaseService_ServiceDesc.ServiceName, healthpb.HealthCheckResponse_NOT_SERVING)
	}

	if a.redisDB != nil {
		if err := a.redisDB.Close(); err != nil {
			lg.Error(ctx, "Failed to close Redis connection", zap.Error(err))
		} else {
			lg.Info(ctx, "Redis connection closed")
		}
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

	lg.Info(ctx, "Shutdown completed successfully")
}
