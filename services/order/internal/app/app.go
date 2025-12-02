package app

import (
	"context"
	pbDispatch "hive/gen/dispatch"
	pbOrder "hive/gen/order"
	pg "hive/pkg/db/postgres"
	"hive/pkg/logger"
	"hive/services/order/internal/config"
	dispatchClient "hive/services/order/internal/infrastructure/client/dispatch"
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
		return nil, err
	}

	repo := repoPostgres.NewPostgresRepo(db.Pool)

	dispatchConn, err := grpc.NewClient(cfg.DispatchAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	dispatcher := dispatchClient.NewDispatchAdapter(pbDispatch.NewDispatchServiceClient(dispatchConn))

	orderService := service.NewOrderService(repo, dispatcher)

	lis, err := net.Listen("tcp", ":"+strconv.Itoa(cfg.GRPCPort))
	if err != nil {
		return nil, err
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
	const op = "app.Run"
	lg := a.lg.With(zap.String("op", op))

	lg.Info(context.Background(), "Running gRPC server", zap.Int("port", a.cfg.GRPCPort))
	if err := a.grpcServer.Serve(a.lis); err != nil {
		errChan <- err
	}
}

func (a *App) Stop(ctx context.Context) {
	const op = "app.Stop"
	lg := a.lg.With(zap.String("op", op))

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
