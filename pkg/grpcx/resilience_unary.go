package grpcx

import (
	"context"
	"hive/pkg/logger"
	"hive/pkg/resilience"
	"strings"
	"sync"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UnaryResilienceConfig struct {
	Retry            resilience.RetryConfig
	Breaker          resilience.BreakerConfig
	ShouldRetry      func(method string) bool
	IsRetryableError func(error) bool
	IsFailure        func(error) bool // для breaker
}

func UnaryResilienceInterceptor(lg logger.Logger, cfg UnaryResilienceConfig) grpc.UnaryClientInterceptor {
	if cfg.ShouldRetry == nil {
		cfg.ShouldRetry = func(_ string) bool { return true }
	}
	if cfg.IsRetryableError == nil {
		cfg.IsRetryableError = defaultRetryable
	}
	if cfg.IsFailure == nil {
		cfg.IsFailure = defaultFailure
	}

	var breakers sync.Map // method -> *resilience.Breaker

	getBreaker := func(method string) *resilience.Breaker {
		if b, ok := breakers.Load(method); ok {
			return b.(*resilience.Breaker)
		}
		name := cfg.Breaker.Name
		if name == "" {
			name = "grpc"
		}
		bc := cfg.Breaker
		bc.Name = name + ":" + method
		br := resilience.NewBreaker(bc, cfg.IsFailure)
		actual, _ := breakers.LoadOrStore(method, br)
		return actual.(*resilience.Breaker)
	}

	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker, opts ...grpc.CallOption,
	) error {
		br := getBreaker(method)

		call := func(c context.Context) error {
			return invoker(c, method, req, reply, cc, opts...)
		}

		return br.Do(ctx, func() error {
			if !cfg.ShouldRetry(method) {
				return call(ctx)
			}
			return resilience.Retry(ctx, cfg.Retry, cfg.IsRetryableError, call)
		})
	}
}

func defaultRetryable(err error) bool {
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
	st, ok := status.FromError(err)
	if !ok {
		return true
	}
	switch st.Code() {
	case codes.Unavailable, codes.DeadlineExceeded, codes.Internal, codes.ResourceExhausted, codes.Aborted:
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

func LogResilienceConfig(lg logger.Logger, name string, r resilience.RetryConfig, b resilience.BreakerConfig) {
	lg.Info(context.Background(), "grpc resilience enabled",
		zap.String("name", name),
		zap.Int("retry_attempts", r.MaxAttempts),
		zap.Duration("retry_base_delay", r.BaseDelay),
		zap.Duration("retry_max_delay", r.MaxDelay),
		zap.Duration("breaker_interval", b.Interval),
		zap.Duration("breaker_timeout", b.Timeout),
	)
}
