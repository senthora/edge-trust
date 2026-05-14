package network

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// Run executes the given operation with network logic.
// It retries based on the provided delays.
// The first attempt is immediate (no delay).
func Run(
	ctx context.Context,
	logger *zap.Logger,
	delays []time.Duration,
	operation func() error,
) error {
	var lastErr error
	for attempt, delay := range delays {
		// wait before network (skip first attempt)
		if attempt > 0 {
			logger.Info(
				"retrying after delay",
				zap.Int("attempt", attempt+1),
				zap.Duration("delay", delay),
			)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}
		// execute operation
		err := operation()
		if err == nil {
			return nil
		}
		lastErr = err

		logger.Warn("attempt failed",
			zap.Int("attempt", attempt+1),
			zap.Error(err),
		)
	}
	// all attempts failed
	logger.Error("all network attempts failed",
		zap.Int("attempts", len(delays)),
		zap.Error(lastErr),
	)
	return fmt.Errorf(
		"network operation failed after %d attempts: %w",
		len(delays),
		lastErr,
	)
}
