// Package checkers provides implementations for various balance checking services.
package checkers

import (
	"chief-checker/internal/config/appConfig"
	"chief-checker/internal/config/serviceConfig"
	"chief-checker/internal/infrastructure/httpClient/client"
	"chief-checker/internal/infrastructure/httpClient/httpConfig"
	"chief-checker/internal/infrastructure/proxyPool"
	"chief-checker/pkg/errors"
	"chief-checker/pkg/proxyManager"
	"chief-checker/pkg/utils"
	"os"
	"time"
)

// InitDebankConfig initializes the configuration for the Debank checker service.
// It sets up all necessary components including HTTP client, proxy pool, and service endpoints.
//
// Parameters:
// - cfg: Debank service settings from the application configuration
//
// Returns:
// - *debankConfig.DebankConfig: initialized configuration
// - error: if initialization fails
func InitDebankConfig(cfg *appConfig.CheckerSettings) (*serviceConfig.ApiCheckerConfig, error) {
	if ok, err := validateDebankParam(cfg); !ok {
		return nil, err
	}

	proxyList, err := initProxies(cfg.ProxyFilePath, cfg.RotateProxy)
	if err != nil {
		return nil, err
	}

	proxyPool, err := initProxyPool(proxyList)
	if err != nil {
		return nil, err
	}
	cfgHttp := httpConfig.Config{
		Timeout:         30 * time.Second,
		MaxRetries:      5,
		RetryDelay:      3 * time.Second,
		UseProxyPool:    cfg.UseProxyPool,
		IsRotatingProxy: cfg.RotateProxy,
		BlockTime:       10 * time.Second,
		BrowserHeaders:  true,
		RandomizeTLS:    true,
		ClientHints:     true,
		SkipHeaders:     map[string]bool{"sec-ch-ua": true, "sec-ch-ua-mobile": true, "sec-ch-ua-platform": true},
		UseUTLS:         false,
		UTLSClientID:    "Chrome_112",
		Servername:      "api.debank.com",
	}

	httpClient := client.NewHttpClient(proxyPool, cfgHttp)

	return &serviceConfig.ApiCheckerConfig{
		BaseURL:         cfg.BaseURL,
		Endpoints:       cfg.Endpoints,
		ContextDeadline: cfg.ContextDeadline,
		HttpClient:      httpClient,
	}, nil
}

// validateDebankParam validates the Debank configuration parameters.
// It checks for required fields and their validity.
//
// Parameters:
// - cfg: Debank service settings to validate
//
// Returns:
// - bool: true if configuration is valid
// - error: description of validation failure
func validateDebankParam(cfg *appConfig.CheckerSettings) (bool, error) {
	if cfg == nil {
		return false, errors.Wrap(errors.ErrValueEmpty, "debank settings is nil")
	}

	if cfg.ProxyFilePath == "" {
		return false, errors.Wrap(errors.ErrValueEmpty, "proxy file path is required")
	}

	if cfg.BaseURL == "" {
		return false, errors.Wrap(errors.ErrValueEmpty, "base URL is required")
	}

	if cfg.ContextDeadline == 0 {
		return false, errors.Wrap(errors.ErrValueEmpty, "context deadline is required")
	}

	if cfg.Endpoints == nil {
		return false, errors.Wrap(errors.ErrValueEmpty, "endpoints are required")
	}

	return true, nil
}

// initProxies reads and parses the proxy list from a file.
// It validates each proxy and returns a list of properly formatted proxy strings.
//
// Parameters:
// - proxyFilePath: path to the file containing proxy list
//
// Returns:
// - []string: list of validated proxy strings
// - error: if reading or parsing fails
func initProxies(proxyFilePath string, rotateProxy bool) ([]string, error) {
	proxylist, err := utils.ReadProxyList(proxyFilePath)
	if err != nil {
		return nil, errors.Wrap(errors.ErrFailedInit, err.Error())
	}

	parser := proxyManager.NewDefaultProxyParser()
	validator := proxyManager.NewDefaultProxyValidator()
	formatter := proxyManager.NewDefaultProxyFormatter()
	manager := proxyManager.NewProxyManager(parser, validator, formatter)

	for _, proxyStr := range proxylist {
		if err := manager.AddProxy(proxyStr); err != nil {
			return nil, errors.Wrap(errors.ErrFailedInit, err.Error())
		}
	}

	proxies := manager.GetProxies()
	proxyStrings := make([]string, len(proxies))
	for i, proxy := range proxies {
		proxyStrings[i] = manager.FormatProxy(proxy)
	}

	if rotateProxy {
		os.Setenv("HTTP_PROXY", proxyStrings[0])
		os.Setenv("HTTPS_PROXY", proxyStrings[0])
	}

	return proxyStrings, nil
}

// initProxyPool creates a proxy pool from a list of proxies in a file.
// The pool manages proxy rotation and availability.
//
// Parameters:
// - proxyFilePath: path to the file containing proxy list
//
// Returns:
// - proxyPool.ProxyPool: initialized proxy pool
// - error: if initialization fails
func initProxyPool(proxyList []string) (proxyPool.ProxyPool, error) {
	proxyPool, err := proxyPool.NewProxyPool(proxyList)
	if err != nil {
		return nil, errors.Wrap(errors.ErrFailedInit, err.Error())
	}
	return proxyPool, nil
}

func InitRabbyConfig(cfg *appConfig.CheckerSettings) (*serviceConfig.ApiCheckerConfig, error) {
	proxyList, err := initProxies(cfg.ProxyFilePath, cfg.RotateProxy)
	if err != nil {
		return nil, err
	}

	proxyPool, err := initProxyPool(proxyList)
	if err != nil {
		return nil, err
	}
	cfgHttp := httpConfig.Config{
		Timeout:         30 * time.Second,
		MaxRetries:      5,
		RetryDelay:      3 * time.Second,
		UseProxyPool:    cfg.UseProxyPool,
		IsRotatingProxy: cfg.RotateProxy,
		BlockTime:       10 * time.Second,
		BrowserHeaders:  true,
		RandomizeTLS:    true,
		ClientHints:     true,
		SkipHeaders:     map[string]bool{"sec-ch-ua": true, "sec-ch-ua-mobile": true, "sec-ch-ua-platform": true},
		UseUTLS:         false,
		UTLSClientID:    "Chrome_112",
		Servername:      "api.rabby.io",
	}

	httpClient := client.NewHttpClient(proxyPool, cfgHttp)

	return &serviceConfig.ApiCheckerConfig{
		BaseURL:         cfg.BaseURL,
		Endpoints:       cfg.Endpoints,
		ContextDeadline: cfg.ContextDeadline,
		HttpClient:      httpClient,
	}, nil
}
