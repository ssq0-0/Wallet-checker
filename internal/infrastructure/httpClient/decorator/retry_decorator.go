package decorator

import (
	"chief-checker/internal/infrastructure/httpClient"
	"context"
	"fmt"
	"time"
)

type retryDecorator struct {
	client     httpClient.HttpClientInterface
	maxRetries int
	retryDelay time.Duration
}

// NewRetryDecorator creates a new retry decorator
func NewRetryDecorator(client httpClient.HttpClientInterface, maxRetries int, retryDelay time.Duration) httpClient.HttpClientInterface {
	return &retryDecorator{
		client:     client,
		maxRetries: maxRetries,
		retryDelay: retryDelay,
	}
}

func (d *retryDecorator) RequestWithRetry(ctx context.Context, url, method string, reqBody, respBody interface{}, headers map[string]string) error {
	var lastErr error
	for i := 0; i < d.maxRetries; i++ {
		err := d.client.SimpleRequest(ctx, url, method, reqBody, respBody, headers)
		if err == nil {
			return nil
		}

		lastErr = err
		if i < d.maxRetries-1 {
			time.Sleep(d.retryDelay)
		}
	}

	return fmt.Errorf("failed after %d retries: %w", d.maxRetries, lastErr)
}

func (d *retryDecorator) SimpleRequest(ctx context.Context, url, method string, reqBody, respBody interface{}, headers map[string]string) error {
	return d.client.SimpleRequest(ctx, url, method, reqBody, respBody, headers)
}

func (d *retryDecorator) GetTimeout() time.Duration {
	return d.client.GetTimeout()
}
