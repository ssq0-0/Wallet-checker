package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"chief-checker/internal/infrastructure/httpClient/httpFactory"
)

// createRequestWithContext builds an HTTP request with the given context, URL, method, and body.
// If reqBody is provided, it's marshaled to JSON and set as the request body.
//
// Parameters:
//   - ctx: Context for request cancellation and timeout
//   - url: Target URL for the request
//   - method: HTTP method (GET, POST, etc.)
//   - reqBody: Optional request body (will be JSON marshaled)
//
// Returns:
//   - The constructed http.Request object and any error that occurred
func (h *HttpClient) createRequestWithContext(ctx context.Context, url, method string, reqBody interface{}) (*http.Request, error) {
	var body io.Reader

	if reqBody != nil {
		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return nil, fmt.Errorf("error marshaling request body: %v", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %v", err)
	}

	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

// executeWithRetries performs the HTTP request with retry logic.
// It will retry on specific error conditions up to the configured maximum number of retries.
// Each retry uses a fresh transport and potentially a new proxy from the pool.
//
// Parameters:
//   - req: The HTTP request to execute
//   - respBody: Pointer to a struct where the response will be unmarshaled
//   - headers: Additional HTTP headers to include with the request
//
// Returns:
//   - Error if all retry attempts fail, nil on success
func (h *HttpClient) executeWithRetries(req *http.Request, respBody interface{}, headers map[string]string) error {
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	ctx := req.Context()
	var currentProxy string

	for attempts := 0; attempts < h.config.MaxRetries; attempts++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// Create a new transport for each request
		transportFactory := httpFactory.NewTransportFactory()
		newTransport := transportFactory.CreateTransport(h.config)

		h.mutex.Lock()
		h.client.Transport = newTransport
		h.transport = newTransport
		h.mutex.Unlock()

		if h.proxyPool != nil && h.config.UseProxyPool {
			currentProxy = h.proxyPool.GetFreeProxy()
			if currentProxy == "" {
				return fmt.Errorf("no available proxies")
			}

			proxyURL, err := url.Parse(currentProxy)
			if err != nil {
				return fmt.Errorf("invalid proxy format %s: %v", currentProxy, err)
			}

			h.mutex.Lock()
			if transport, ok := h.transport.(*http.Transport); ok {
				transport.Proxy = http.ProxyURL(proxyURL)
			}
			h.mutex.Unlock()
		}

		resp, err := h.client.Do(req)

		// Close all connections after the request
		if transport, ok := h.client.Transport.(*http.Transport); ok {
			transport.CloseIdleConnections()
		}

		if shouldRetry, retryErr := h.handleRequestError(err, currentProxy); shouldRetry {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(h.config.RetryDelay):
				continue
			}
		} else if retryErr != nil {
			return retryErr
		}

		if resp == nil {
			return fmt.Errorf("received empty response")
		}
		defer resp.Body.Close()

		if shouldRetry, retryErr := h.handleResponseStatus(resp, currentProxy, attempts); shouldRetry {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(h.config.RetryDelay):
				continue
			}
		} else if retryErr != nil {
			return retryErr
		}

		if err := h.parseResponse(resp, respBody); err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("request failed after %d attempts", h.config.MaxRetries)
}

// executeWithoutRetries performs the HTTP request without retry logic.
// It still uses proxy rotation if configured, but doesn't retry on failures.
//
// Parameters:
//   - req: The HTTP request to execute
//   - respBody: Pointer to a struct where the response will be unmarshaled
//   - headers: Additional HTTP headers to include with the request
//
// Returns:
//   - Error if the request fails, nil on success
func (h *HttpClient) executeWithoutRetries(req *http.Request, respBody interface{}, headers map[string]string) error {
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	ctx := req.Context()
	var currentProxy string

	// Create a new transport for each request
	transportFactory := httpFactory.NewTransportFactory()
	newTransport := transportFactory.CreateTransport(h.config)

	h.mutex.Lock()
	h.client.Transport = newTransport
	h.transport = newTransport
	h.mutex.Unlock()

	if h.proxyPool != nil && h.config.UseProxyPool {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		currentProxy = h.proxyPool.GetFreeProxy()
		if currentProxy != "" {
			proxyURL, err := url.Parse(currentProxy)
			if err != nil {
				return fmt.Errorf("invalid proxy format %s: %v", currentProxy, err)
			}

			h.mutex.Lock()
			if transport, ok := h.transport.(*http.Transport); ok {
				transport.Proxy = http.ProxyURL(proxyURL)
			}
			h.mutex.Unlock()
		}
	}

	resp, err := h.client.Do(req)

	// Close all connections after the request
	if transport, ok := h.client.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}

	if err != nil {
		if h.proxyPool != nil && currentProxy != "" && !h.config.IsRotatingProxy &&
			(strings.Contains(err.Error(), "Proxy Authentication Required") ||
				strings.Contains(err.Error(), "authentication failed") ||
				strings.Contains(err.Error(), "context deadline exceeded")) {

			h.proxyPool.BlockProxy(currentProxy, h.config.BlockTime)
		}
		return fmt.Errorf("request error: %v", err)
	}

	if resp == nil {
		return fmt.Errorf("received empty response")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusProxyAuthRequired {
		if h.proxyPool != nil && currentProxy != "" && !h.config.IsRotatingProxy {
			h.proxyPool.BlockProxy(currentProxy, h.config.BlockTime)
		}
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("status error: %d (rate limit or proxy issue): %s", resp.StatusCode, string(body))
	}

	return h.parseResponse(resp, respBody)
}
