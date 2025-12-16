package grpcx

import (
	"context"

	"google.golang.org/grpc"
)

func ChainUnaryClientInterceptors(inters ...grpc.UnaryClientInterceptor) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker, opts ...grpc.CallOption,
	) error {
		chained := invoker
		for i := len(inters) - 1; i >= 0; i-- {
			ii := inters[i]
			next := chained
			chained = func(c context.Context, m string, r, rep any, conn *grpc.ClientConn, o ...grpc.CallOption) error {
				return ii(c, m, r, rep, conn, next, o...)
			}
		}
		return chained(ctx, method, req, reply, cc, opts...)
	}
}
