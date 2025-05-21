package client

import (
	"bytes"
	"chief-checker/internal/infrastructure/httpClient"
	"chief-checker/internal/infrastructure/httpClient/config"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type baseClient struct {
	client  *http.Client
	timeout time.Duration
}

// NewBaseClient creates a new base HTTP client
func NewBaseClient(cfg *config.Config) (httpClient.HttpClientInterface, error) {
	// Настраиваем транспорт по аналогии с рабочей версией: отключаем HTTP/2 и Keep-Alives
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
			NextProtos:         []string{"http/1.1"},
		},
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		MaxConnsPerHost:     100,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		DisableKeepAlives:   true,
		ForceAttemptHTTP2:   false,
		TLSNextProto:        make(map[string]func(string, *tls.Conn) http.RoundTripper),
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   cfg.Timeout,
	}

	return &baseClient{
		client:  client,
		timeout: cfg.Timeout,
	}, nil
}

func (c *baseClient) RequestWithRetry(ctx context.Context, url, method string, reqBody, respBody interface{}, headers map[string]string) error {
	return c.SimpleRequest(ctx, url, method, reqBody, respBody, headers)
}

func (c *baseClient) SimpleRequest(ctx context.Context, url, method string, reqBody, respBody interface{}, headers map[string]string) error {
	var bodyReader io.Reader
	if reqBody != nil {
		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if respBody != nil {
		if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
			return fmt.Errorf("failed to decode response body: %w", err)
		}
	}

	return nil
}

func (c *baseClient) GetTimeout() time.Duration {
	return c.timeout
}
