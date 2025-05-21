package factory

import (
	"chief-checker/internal/infrastructure/httpClient"
	"chief-checker/internal/infrastructure/httpClient/client"
	"chief-checker/internal/infrastructure/httpClient/config"
	"chief-checker/internal/infrastructure/httpClient/decorator"
	"chief-checker/internal/infrastructure/proxyPool"
)

// NewHttpClient creates a new HTTP client with the given configuration
func NewHttpClient(cfg *config.Config, proxyPool proxyPool.ProxyPool) (httpClient.HttpClientInterface, error) {
	baseClient, err := client.NewBaseClient(cfg)
	if err != nil {
		return nil, err
	}

	// Apply decorators based on configuration
	var client httpClient.HttpClientInterface = baseClient

	if cfg.UseProxyPool && proxyPool != nil {
		client = decorator.NewProxyDecorator(client, proxyPool)
	}

	if cfg.UseUTLS {
		client = decorator.NewUTLSDecorator(client, cfg.UTLSClientID)
	}

	if cfg.MaxRetries > 0 {
		client = decorator.NewRetryDecorator(client, cfg.MaxRetries, cfg.RetryDelay)
	}

	return client, nil
}
