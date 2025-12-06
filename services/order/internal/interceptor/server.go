package interceptor

import (
	"context"
	"hive/pkg/logger"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func LoggingUnaryServerInterceptor(lg logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		lg = lg.With(
			zap.String("component", "grpc_server_interceptor"),
			zap.String("method", info.FullMethod),
		)

		lg.Info(context.Background(), "gRPC server request started", zap.Any("server", info.Server), zap.Any("req", req))
		start := time.Now()
		resp, err := handler(ctx, req)
		dur := time.Since(start)
		if err != nil {
			lg.Error(context.Background(), "gRPC server request failed",
				zap.Duration("duration", dur),
				zap.Error(err),
			)
		} else {
			lg.Info(context.Background(), "gRPC server request completed",
				zap.Duration("duration", dur),
			)
		}

		return resp, err
	}
}
