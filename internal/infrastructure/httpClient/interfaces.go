package httpClient

import (
	"context"
	"time"
)

// HttpClientInterface defines the contract for an HTTP client
type HttpClientInterface interface {
	// RequestWithRetry sends an HTTP request with automatic retries on failures
	RequestWithRetry(ctx context.Context, url, method string, reqBody, respBody interface{}, headers map[string]string) error

	// SimpleRequest sends an HTTP request without retry mechanism
	SimpleRequest(ctx context.Context, url, method string, reqBody, respBody interface{}, headers map[string]string) error

	// GetTimeout returns the client's timeout setting
	GetTimeout() time.Duration
}
