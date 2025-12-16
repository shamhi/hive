package grpcx

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"hive/pkg/logger"
	"hive/pkg/resilience"

	"github.com/sony/gobreaker/v2"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ClientResilienceConfig struct {
	Name string

	Timeout time.Duration

	Retry   resilience.RetryConfig
	Breaker resilience.BreakerConfig

	ShouldRetryMethod func(method string) bool
	IsRetryableError  func(error) bool
	IsFailure         func(error) bool
}

func UnaryClientResilienceInterceptor(lg logger.Logger, cfg ClientResilienceConfig) grpc.UnaryClientInterceptor {
	if cfg.Name == "" {
		cfg.Name = "grpc_client"
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 2 * time.Second
	}
	if cfg.ShouldRetryMethod == nil {
		cfg.ShouldRetryMethod = RetryOnlyReads
	}
	if cfg.IsRetryableError == nil {
		cfg.IsRetryableError = defaultRetryableError
	}
	if cfg.IsFailure == nil {
		cfg.IsFailure = defaultFailure
	}

	var breakers sync.Map
	getBreaker := func(method string) *resilience.Breaker {
		if b, ok := breakers.Load(method); ok {
			return b.(*resilience.Breaker)
		}
		bc := cfg.Breaker
		if bc.Name == "" {
			bc.Name = cfg.Name
		}
		bc.Name = bc.Name + ":" + method
		br := resilience.NewBreaker(bc, cfg.IsFailure)
		actual, _ := breakers.LoadOrStore(method, br)
		return actual.(*resilience.Breaker)
	}

	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		l := lg.With(
			zap.String("component", "grpc_client"),
			zap.String("method", method),
		)

		ctx, cancel := withMaxTimeout(ctx, cfg.Timeout)
		defer cancel()

		start := time.Now()
		attempts := 0

		call := func(c context.Context) error {
			attempts++
			return invoker(c, method, req, reply, cc, opts...)
		}

		br := getBreaker(method)

		runOnce := func(c context.Context) error {
			err := br.Do(c, func() error { return call(c) })
			if errors.Is(err, gobreaker.ErrOpenState) || errors.Is(err, gobreaker.ErrTooManyRequests) {
				return status.Error(codes.Unavailable, "upstream temporarily unavailable")
			}
			return err
		}

		l.Info(ctx, "gRPC request started", zap.Any("req", req))

		var err error
		if cfg.ShouldRetryMethod(method) {
			err = resilience.Retry(ctx, cfg.Retry, cfg.IsRetryableError, runOnce)
		} else {
			err = runOnce(ctx)
		}

		dur := time.Since(start)
		if err != nil {
			l.Error(ctx, "gRPC request failed",
				zap.Duration("duration", dur),
				zap.Int("attempts", attempts),
				zap.Error(err),
			)
		} else {
			l.Info(ctx, "gRPC request completed",
				zap.Duration("duration", dur),
				zap.Int("attempts", attempts),
			)
		}

		return err
	}
}

func UnaryServerLoggingTimeoutInterceptor(lg logger.Logger, maxTimeout time.Duration) grpc.UnaryServerInterceptor {
	if maxTimeout <= 0 {
		maxTimeout = 10 * time.Second
	}

	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		l := lg.With(
			zap.String("component", "grpc_server"),
			zap.String("method", info.FullMethod),
		)

		ctx, cancel := withMaxTimeout(ctx, maxTimeout)
		defer cancel()

		start := time.Now()
		resp, err := handler(ctx, req)
		dur := time.Since(start)

		if err != nil {
			l.Error(ctx, "gRPC request failed", zap.Duration("duration", dur), zap.Error(err))
		} else {
			l.Info(ctx, "gRPC request completed", zap.Duration("duration", dur))
		}
		return resp, err
	}
}

func LoggingTimeoutStreamServerInterceptor(lg logger.Logger, maxTimeout time.Duration) grpc.StreamServerInterceptor {
	if maxTimeout <= 0 {
		maxTimeout = 10 * time.Minute
	}

	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		l := lg.With(
			zap.String("component", "grpc_stream_server"),
			zap.String("method", info.FullMethod),
		)

		ctx := ss.Context()
		ctx, cancel := withMaxTimeout(ctx, maxTimeout)
		defer cancel()

		wrapped := &serverStreamWithContext{ServerStream: ss, ctx: ctx}

		start := time.Now()
		l.Info(ctx, "gRPC stream started", zap.Any("server", srv))

		err := handler(srv, wrapped)

		dur := time.Since(start)
		if err != nil {
			st, _ := status.FromError(err)
			l.Error(ctx, "gRPC stream failed",
				zap.Duration("duration", dur),
				zap.String("code", st.Code().String()),
				zap.Error(err),
			)
		} else {
			l.Info(ctx, "gRPC stream completed", zap.Duration("duration", dur))
		}
		return err
	}
}

type serverStreamWithContext struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *serverStreamWithContext) Context() context.Context {
	return s.ctx
}

func withMaxTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		return ctx, func() {}
	}
	if dl, ok := ctx.Deadline(); ok {
		if time.Until(dl) <= timeout {
			return ctx, func() {}
		}
	}
	return context.WithTimeout(ctx, timeout)
}

func defaultRetryableError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	if errors.Is(err, gobreaker.ErrOpenState) || errors.Is(err, gobreaker.ErrTooManyRequests) {
		return false
	}

	st, ok := status.FromError(err)
	if !ok {
		return true
	}

	switch st.Code() {
	case codes.Unavailable, codes.DeadlineExceeded, codes.ResourceExhausted, codes.Aborted:
		return true
	default:
		return false
	}
}

func defaultFailure(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	st, ok := status.FromError(err)
	if !ok {
		return true
	}

	switch st.Code() {
	case codes.Unavailable, codes.ResourceExhausted, codes.Internal:
		return true
	default:
		return false
	}
}

func RetryOnlyReads(method string) bool {
	parts := strings.Split(method, "/")
	m := parts[len(parts)-1]
	return strings.HasPrefix(m, "Get") || strings.HasPrefix(m, "Find") || strings.HasPrefix(m, "List")
}
