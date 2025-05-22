// Package api provides functionality for interacting with external APIs
// with built-in retry mechanisms and error handling.
package api

import (
	"chief-checker/internal/usecase/features/checkerUsecase/types"
	"chief-checker/pkg/logger"
	"context"
	"fmt"
	"math/rand"
	"time"
)

// RetryableOperation represents a function that can be retried.
// It returns a result of any type and an error.
type RetryableOperation func() (interface{}, error)

// retryFunc executes an operation with automatic retries on failure.
// It implements exponential backoff with jitter for retry delays.
//
// Parameters:
// - ctx: context for cancellation
// - operation: name of the operation for logging
// - address: address being processed
// - fn: function to retry
//
// Returns:
// - interface{}: result of the successful operation
// - error: if all retries fail or context is cancelled
func (c *DataCollector) retryFunc(ctx context.Context, operation string, address string, fn RetryableOperation) (interface{}, error) {
	var lastError error

	for attempt := 0; attempt < types.MaxRetries; attempt++ {
		if attempt > 0 {
			c.errorCollector.SaveError(address, fmt.Sprintf("Retry attempt %d/%d for %s", attempt, types.MaxRetries, operation))
			select {
			case <-ctx.Done():
				c.errorCollector.SaveError(address, ctx.Err().Error())
				return nil, ctx.Err()
			case <-time.After(time.Duration(types.RetryDelayBase+rand.Intn(types.RetryDelayRandom)) * time.Millisecond):
			}
		}

		result, err := fn()
		if err == nil {
			return result, nil
		}

		lastError = err
		c.errorCollector.SaveError(address, fmt.Sprintf("%s failed (attempt %d/%d): %v", operation, attempt+1, types.MaxRetries, err))
	}

	logger.GlobalLogger.Errorf("[%s] %s failed after %d attempts: %v", address, operation, types.MaxRetries, lastError)
	return nil, lastError
}
