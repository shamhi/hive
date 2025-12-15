package interceptor

import (
	"context"
	"hive/pkg/logger"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func TimeoutUnaryClientInterceptor(lg logger.Logger, timeout time.Duration) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		lg := lg.With(
			zap.String("component", "grpc_client_interceptor"),
			zap.String("method", method),
		)

		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		lg.Info(ctx, "gRPC client request started", zap.Any("req", req))
		start := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		dur := time.Since(start)
		if err != nil {
			lg.Error(ctx, "gRPC client request failed",
				zap.Duration("duration", dur),
				zap.Error(err),
			)
		} else {
			lg.Info(ctx, "gRPC client request completed",
				zap.Duration("duration", dur),
			)
		}

		return err
	}
}
