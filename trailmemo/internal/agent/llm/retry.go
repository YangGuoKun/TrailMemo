package llm

import (
	"context"
	"errors"
	"math"
	"net"
	"net/http"
	"time"

	"github.com/trailmemo/internal/platform/logger"
	"go.uber.org/zap"
)

// RetryConfig controls retry behaviour for LLM client calls.
type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
}

// DefaultRetryConfig is the recommended retry policy for LLM calls.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: 2,
		BaseDelay:  1 * time.Second,
		MaxDelay:   8 * time.Second,
	}
}

// isRetryable returns true for transient errors where a retry might succeed.
func isRetryable(err error) bool {
	if err == nil {
		return false
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return false // context timeout means the caller gave up
	}
	return true
}

func statusIsRetryable(status int) bool {
	return status == http.StatusTooManyRequests ||
		status == http.StatusInternalServerError ||
		status == http.StatusBadGateway ||
		status == http.StatusServiceUnavailable ||
		status == http.StatusGatewayTimeout
}

// Retry executes fn with exponential backoff on retryable errors.
func Retry[T any](ctx context.Context, cfg RetryConfig, opName string, fn func(context.Context) (T, error)) (T, error) {
	var lastErr error
	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		result, err := fn(ctx)
		if err == nil {
			return result, nil
		}
		lastErr = err

		if attempt == cfg.MaxRetries || !isRetryable(err) {
			break
		}

		delay := time.Duration(math.Min(
			float64(cfg.BaseDelay)*math.Pow(2, float64(attempt)),
			float64(cfg.MaxDelay),
		))

		logger.FromContext(ctx).Warn("llm_retry",
			zap.String("operation", opName),
			zap.Int("attempt", attempt+1),
			zap.Int("max_retries", cfg.MaxRetries),
			zap.Duration("delay", delay),
			zap.Error(err),
		)

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return result, ctx.Err()
		case <-timer.C:
		}
	}

	logger.FromContext(ctx).Error("llm_retry_exhausted",
		zap.String("operation", opName),
		zap.Int("max_retries", cfg.MaxRetries),
		zap.Error(lastErr),
	)
	var zero T
	return zero, lastErr
}
