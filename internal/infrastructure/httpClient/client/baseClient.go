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

type HttpClient struct {
	client    *http.Client
	config    *httpConfig.Config
	proxyPool proxyPool.ProxyPool
	transport http.RoundTripper
	mutex     sync.Mutex
}

// NewHttpClient creates a new HTTP client instance with the specified proxy pool.
// It configures the transport layer with proper TLS and proxy settings.
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

func (h *HttpClient) RequestWithRetry(ctx context.Context, urlStr, method string, reqBody, respBody interface{}, headers map[string]string) error {
	req, err := h.createRequestWithContext(ctx, urlStr, method, reqBody)
	if err != nil {
		return fmt.Errorf("ошибка создания запроса: %v", err)
	}

	return h.executeWithRetries(req, respBody, headers)
}

func (h *HttpClient) SimpleRequest(ctx context.Context, urlStr, method string, reqBody, respBody interface{}, headers map[string]string) error {
	req, err := h.createRequestWithContext(ctx, urlStr, method, reqBody)
	if err != nil {
		return fmt.Errorf("ошибка создания запроса: %v", err)
	}

	return h.executeWithoutRetries(req, respBody, headers)
}

// GetTimeout возвращает таймаут HTTP клиента
func (h *HttpClient) GetTimeout() time.Duration {
	return h.client.Timeout
}
