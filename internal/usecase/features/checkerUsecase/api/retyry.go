package api

import (
	"chief-checker/internal/usecase/features/checkerUsecase/types"
	"chief-checker/pkg/logger"
	"context"
	"fmt"
	"math/rand"
	"time"
)

type RetryableOperation func() (interface{}, error)

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
