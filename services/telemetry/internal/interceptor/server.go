package interceptor

import (
	"hive/pkg/logger"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func LoggingStreamServerInterceptor(lg logger.Logger) grpc.StreamServerInterceptor {
	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		lg := lg.With(
			zap.String("component", "grpc_server_interceptor"),
			zap.String("method", info.FullMethod),
		)

		lg.Info(ss.Context(), "gRPC server request started", zap.Any("server", srv))
		start := time.Now()
		err := handler(srv, ss)
		dur := time.Since(start)
		if err != nil {
			lg.Error(ss.Context(), "gRPC server request failed",
				zap.Duration("duration", dur),
				zap.Error(err),
			)
		} else {
			lg.Info(ss.Context(), "gRPC server request completed",
				zap.Duration("duration", dur),
			)
		}

		return err
	}
}
