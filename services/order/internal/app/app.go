package app

import (
	"context"
	"fmt"
	pbDispatch "hive/gen/dispatch"
	pbOrder "hive/gen/order"
	pg "hive/pkg/db/postgres"
	"hive/pkg/logger"
	"hive/services/order/internal/config"
	grpcDispatchClient "hive/services/order/internal/infrastructure/client/dispatch"
	repoPostgres "hive/services/order/internal/repository/postgres"
	"hive/services/order/internal/service"
	transportGrpc "hive/services/order/internal/transport/grpc"
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
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}

	repo := repoPostgres.NewPostgresRepo(db.Pool)

	dispatchConn, err := grpc.NewClient(
		cfg.DispatchAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to dispatch service: %w", err)
	}
	dispatchClient := grpcDispatchClient.NewDispatchClient(pbDispatch.NewDispatchServiceClient(dispatchConn))

	orderService := service.NewOrderService(
		repo,
		dispatchClient,
	)

	lis, err := net.Listen("tcp", ":"+strconv.Itoa(cfg.GRPCPort))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %w", cfg.GRPCPort, err)
	}

	grpcServer := grpc.NewServer()
	orderServer := transportGrpc.NewServer(orderService)
	pbOrder.RegisterOrderServiceServer(grpcServer, orderServer)

	return &App{
		cfg:        cfg,
		lg:         lg,
		lis:        lis,
		grpcServer: grpcServer,
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

	lg.Info(ctx, "Shutting down...")

	done := make(chan struct{})
	go func() {
		a.grpcServer.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		lg.Info(ctx, "gRPC server gracefully stopped")
	case <-ctx.Done():
		lg.Warn(ctx, "Timeout reached, force stopping...")
		a.grpcServer.Stop()
		lg.Info(ctx, "gRPC server force stopped")
	}
}
