package client

import (
	"chief-checker/pkg/logger"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// handleRequestError processes HTTP request errors and determines if a retry is needed.
// If the error is related to proxy authentication or timeouts, it will block the proxy
// and signal that a retry should be attempted.
//
// Parameters:
//   - err: The error returned from the HTTP request
//   - currentProxy: The proxy URL used for the request (if any)
//
// Returns:
//   - bool: True if retry is recommended, false otherwise
//   - error: The error to return if no retry should be attempted, nil otherwise
func (h *HttpClient) handleRequestError(err error, currentProxy string) (bool, error) {
	if err == nil {
		return false, nil
	}

	if strings.Contains(err.Error(), "Proxy Authentication Required") ||
		strings.Contains(err.Error(), "authentication failed") ||
		strings.Contains(err.Error(), "context deadline exceeded") {

		if h.proxyPool != nil && !h.config.IsRotatingProxy {
			h.proxyPool.BlockProxy(currentProxy, h.config.BlockTime)
		}

		return true, nil
	}

	return false, fmt.Errorf("request error: %v", err)
}

// handleResponseStatus processes response statuses and determines if a retry is needed.
// It handles rate limiting (429) and proxy authentication issues (407) by blocking
// the current proxy and recommending a retry.
//
// Parameters:
//   - resp: The HTTP response to process
//   - currentProxy: The proxy URL used for the request (if any)
//   - attempt: The current retry attempt number (0-indexed)
//
// Returns:
//   - bool: True if retry is recommended, false otherwise
//   - error: The error to return if no retry should be attempted, nil otherwise
func (h *HttpClient) handleResponseStatus(resp *http.Response, currentProxy string, attempt int) (bool, error) {
	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusProxyAuthRequired {
		logger.GlobalLogger.Debugf("Rate limit/proxy issues. Attempt %d", attempt+1)

		if h.proxyPool != nil && !h.config.IsRotatingProxy {
			h.proxyPool.BlockProxy(currentProxy, h.config.BlockTime)
		}

		return true, nil
	}

	return false, nil
}

// parseResponse handles the HTTP response, including gzip decompression
// and JSON unmarshaling into the provided response body structure.
//
// Parameters:
//   - resp: The HTTP response to process
//   - respBody: Pointer to a struct where the response will be unmarshaled
//
// Returns:
//   - Error if parsing fails, nil on success
func (h *HttpClient) parseResponse(resp *http.Response, respBody interface{}) error {
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.GlobalLogger.Debugf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	reader, err := h.getResponseReader(resp)
	if err != nil {
		return err
	}
	defer reader.Close()

	body, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("error reading response body: %v", err)
	}

	if respBody == nil {
		return nil
	}

	if err := json.Unmarshal(body, respBody); err != nil {
		return fmt.Errorf("error parsing JSON response: %v", err)
	}

	return nil
}

// getResponseReader returns the appropriate reader based on the Content-Encoding header.
// It handles gzip-compressed responses by creating a gzip reader.
//
// Parameters:
//   - resp: The HTTP response
//
// Returns:
//   - An io.ReadCloser for reading the response body
//   - Error if creating the reader fails
func (h *HttpClient) getResponseReader(resp *http.Response) (io.ReadCloser, error) {
	if resp.Header.Get("Content-Encoding") != "gzip" {
		return resp.Body, nil
	}

	gzReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %v", err)
	}

	return gzReader, nil
}
