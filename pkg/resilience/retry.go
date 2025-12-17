package resilience

import (
	"context"
	"math"
	"math/rand"
	"time"
)

func init() { rand.Seed(time.Now().UnixNano()) }

type RetryConfig struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	Jitter      float64
}

func Retry(ctx context.Context, cfg RetryConfig, shouldRetry func(error) bool, fn func(context.Context) error) error {
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 1
	}
	if cfg.BaseDelay <= 0 {
		cfg.BaseDelay = 50 * time.Millisecond
	}
	if cfg.MaxDelay <= 0 {
		cfg.MaxDelay = 1 * time.Second
	}
	if cfg.Jitter < 0 {
		cfg.Jitter = 0
	}
	if cfg.Jitter > 1 {
		cfg.Jitter = 1
	}
	if shouldRetry == nil {
		shouldRetry = func(error) bool { return true }
	}

	var lastErr error
	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		err := fn(ctx)
		if err == nil {
			return nil
		}
		lastErr = err

		if attempt == cfg.MaxAttempts || !shouldRetry(err) {
			return lastErr
		}

		delay := backoff(cfg, attempt)
		t := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			t.Stop()
			return ctx.Err()
		case <-t.C:
		}
	}
	return lastErr
}

func backoff(cfg RetryConfig, attempt int) time.Duration {
	pow := math.Pow(2, float64(attempt-1))
	d := time.Duration(float64(cfg.BaseDelay) * pow)
	if d > cfg.MaxDelay {
		d = cfg.MaxDelay
	}
	if cfg.Jitter > 0 {
		amp := float64(d) * cfg.Jitter
		delta := (rand.Float64()*2 - 1) * amp
		d = time.Duration(float64(d) + delta)
		if d < 0 {
			d = 0
		}
	}
	return d
}
