package checkers

import (
	"chief-checker/internal/config/appConfig"
	"chief-checker/internal/config/serviceConfig/debankConfig"
	"chief-checker/internal/infrastructure/httpClient/client"
	"chief-checker/internal/infrastructure/httpClient/httpConfig"
	"chief-checker/internal/infrastructure/proxyPool"
	"chief-checker/pkg/errors"
	"chief-checker/pkg/proxyManager"
	"chief-checker/pkg/utils"
	"time"
)

// InitDebankConfig инициализирует конфиг для Debank чекера.
func InitDebankConfig(cfg *appConfig.DebankSettings) (*debankConfig.DebankConfig, error) {
	if ok, err := validateDebankParam(cfg); !ok {
		return nil, err
	}

	proxyPool, err := initProxyPool(cfg.ProxyFilePath)
	if err != nil {
		return nil, err
	}
	cfgHttp := httpConfig.Config{
		Timeout:         10 * time.Second,
		MaxRetries:      5,
		RetryDelay:      3 * time.Second,
		UseProxyPool:    true,
		IsRotatingProxy: true,
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

	return &debankConfig.DebankConfig{
		BaseURL:         cfg.BaseURL,
		Endpoints:       cfg.Endpoints,
		ContextDeadline: cfg.ContextDeadline,
		HttpClient:      httpClient,
	}, nil
}

// validateDebankParam валидирует параметры конфига Debank.
func validateDebankParam(cfg *appConfig.DebankSettings) (bool, error) {
	if cfg == nil {
		return false, errors.Wrap(errors.ErrValueEmpty, "debank settings is nil")
	}

	if cfg.ProxyFilePath == "" ||
		cfg.BaseURL == "" ||
		cfg.ContextDeadline == 0 ||
		cfg.Endpoints == nil {
		return false, errors.Wrap(errors.ErrValueEmpty, "proxy file path argument not find")
	}
	return true, nil
}

func initProxies(proxyFilePath string) ([]string, error) {
	proxylist, err := utils.ReadProxyList(proxyFilePath)
	if err != nil {
		return nil, errors.Wrap(errors.ErrFailedInit, err.Error())
	}

	proxyManager, err := proxyManager.NewProxyManager(proxylist)
	if err != nil {
		return nil, errors.Wrap(errors.ErrFailedInit, err.Error())
	}

	proxies, err := proxyManager.ParseProxy(proxylist)
	if err != nil {
		return nil, errors.Wrap(errors.ErrFailedInit, err.Error())
	}

	return proxies, nil
}

// initProxyPool создает пул прокси из файла.
func initProxyPool(proxyFilePath string) (proxyPool.ProxyPool, error) {
	proxylist, err := utils.ReadProxyList(proxyFilePath)
	if err != nil {
		return nil, errors.Wrap(errors.ErrFailedInit, err.Error())
	}

	proxyPool, err := proxyPool.NewProxyPool(proxylist)
	if err != nil {
		return nil, errors.Wrap(errors.ErrFailedInit, err.Error())
	}
	return proxyPool, nil
}
