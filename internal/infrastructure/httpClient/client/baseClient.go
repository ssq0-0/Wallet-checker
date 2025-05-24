// Package client provides a robust HTTP client implementation with advanced features.
//
// It includes:
// - Configurable retry logic with exponential backoff
// - Automatic proxy rotation and management
// - TLS fingerprint customization
// - Response handling and parsing
// - Thread-safe transport management
//
// This package is designed to work with anti-bot protection systems and
// provides tools for reliable HTTP communication in challenging environments.
package client

import (
	"chief-checker/internal/infrastructure/httpClient/httpConfig"
	"chief-checker/internal/infrastructure/httpClient/httpFactory"
	"chief-checker/internal/infrastructure/httpClient/httpInterfaces"
	"chief-checker/internal/infrastructure/proxyPool"
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// HttpClient implements the httpInterfaces.HttpClientInterface and provides
// a high-level HTTP client with advanced functionality.
// It manages an underlying http.Client with configurable transport, proxies, and retry logic.
type HttpClient struct {
	client    *http.Client        // Standard Go HTTP client
	config    *httpConfig.Config  // Configuration for the client
	proxyPool proxyPool.ProxyPool // Proxy management interface
	transport http.RoundTripper   // Transport for HTTP requests
	mutex     sync.Mutex          // Mutex for thread-safe operations
}

// NewHttpClient creates a new HTTP client instance with the specified proxy pool.
// It configures the transport layer with proper TLS and proxy settings.
//
// Parameters:
//   - proxyPool: ProxyPool interface for managing and rotating proxies
//   - config: Configuration settings for the HTTP client
//
// Returns:
//   - An implementation of httpInterfaces.HttpClientInterface
func NewHttpClient(proxyPool proxyPool.ProxyPool, config httpConfig.Config) httpInterfaces.HttpClientInterface {
	transportFactory := httpFactory.NewTransportFactory()
	transport := transportFactory.CreateTransport(&config)
	client := &HttpClient{
		client:    &http.Client{Transport: transport, Timeout: config.Timeout},
		transport: transport,
		proxyPool: proxyPool,
		config:    &config,
	}

	return client
}

// RequestWithRetry performs an HTTP request with automatic retries on failures.
// It will retry on network errors, proxy authentication issues, and rate limiting.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - urlStr: The URL to request
//   - method: The HTTP method (GET, POST, etc.)
//   - reqBody: Request body to be marshaled as JSON (can be nil)
//   - respBody: Pointer to a struct where the response will be unmarshaled
//   - headers: Additional HTTP headers to include with the request
//
// Returns:
//   - Error if all retry attempts fail, nil on success
func (h *HttpClient) RequestWithRetry(ctx context.Context, urlStr, method string, reqBody, respBody interface{}, headers map[string]string) error {
	req, err := h.createRequestWithContext(ctx, urlStr, method, reqBody)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	return h.executeWithRetries(req, respBody, headers)
}

// SimpleRequest performs a single HTTP request without retry logic.
// This is useful for non-critical requests or when custom retry logic is needed.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - urlStr: The URL to request
//   - method: The HTTP method (GET, POST, etc.)
//   - reqBody: Request body to be marshaled as JSON (can be nil)
//   - respBody: Pointer to a struct where the response will be unmarshaled
//   - headers: Additional HTTP headers to include with the request
//
// Returns:
//   - Error if the request fails, nil on success
func (h *HttpClient) SimpleRequest(ctx context.Context, urlStr, method string, reqBody, respBody interface{}, headers map[string]string) error {
	req, err := h.createRequestWithContext(ctx, urlStr, method, reqBody)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	return h.executeWithoutRetries(req, respBody, headers)
}

// GetTimeout returns the current timeout setting for the HTTP client.
//
// Returns:
//   - The current timeout duration
func (h *HttpClient) GetTimeout() time.Duration {
	return h.client.Timeout
}
