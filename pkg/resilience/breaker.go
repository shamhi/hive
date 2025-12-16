package resilience

import (
	"context"
	"time"

	"github.com/sony/gobreaker/v2"
)

type BreakerConfig struct {
	Name        string
	Interval    time.Duration // окно статистики
	Timeout     time.Duration // как долго держим open до half-open
	MaxRequests uint32        // в half-open
	MinRequests uint32        // минимум запросов для оценки
	FailureRate float64       // 0..1
}

type Breaker struct {
	cb *gobreaker.CircuitBreaker[struct{}]
}

func NewBreaker(cfg BreakerConfig, isFailure func(error) bool) *Breaker {
	if cfg.Interval <= 0 {
		cfg.Interval = 10 * time.Second
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 5 * time.Second
	}
	if cfg.MaxRequests == 0 {
		cfg.MaxRequests = 3
	}
	if cfg.MinRequests == 0 {
		cfg.MinRequests = 5
	}
	if cfg.FailureRate <= 0 || cfg.FailureRate > 1 {
		cfg.FailureRate = 0.6
	}
	if isFailure == nil {
		isFailure = func(err error) bool { return true }
	}

	settings := gobreaker.Settings{
		Name:        cfg.Name,
		Interval:    cfg.Interval,
		Timeout:     cfg.Timeout,
		MaxRequests: cfg.MaxRequests,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			if counts.Requests < cfg.MinRequests {
				return false
			}
			rate := float64(counts.TotalFailures) / float64(counts.Requests)
			return rate >= cfg.FailureRate
		},
		IsSuccessful: func(err error) bool {
			return err == nil || !isFailure(err)
		},
	}

	return &Breaker{cb: gobreaker.NewCircuitBreaker[struct{}](settings)}
}

func (b *Breaker) Do(ctx context.Context, fn func() error) error {
	_, err := b.cb.Execute(func() (struct{}, error) {
		if err := ctx.Err(); err != nil {
			return struct{}{}, err
		}
		return struct{}{}, fn()
	})
	return err
}
