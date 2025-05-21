package checkers

import (
	"chief-checker/internal/config/appConfig"
	"chief-checker/internal/config/serviceConfig/debankConfig"
	"chief-checker/internal/infrastructure/httpClient"
	httpConfig "chief-checker/internal/infrastructure/httpClient/config"
	httpFactory "chief-checker/internal/infrastructure/httpClient/factory"
	"chief-checker/internal/infrastructure/proxyPool"
	"chief-checker/pkg/errors"
	"chief-checker/pkg/proxyManager"
	"chief-checker/pkg/utils"
)

func InitDebankConfig(cfg *appConfig.DebankSettings) (*debankConfig.DebankConfig, error) {
	if ok, err := validateDebankParam(cfg); !ok {
		return nil, err
	}

	proxyPool, err := initProxyPool(cfg.ProxyFilePath)
	if err != nil {
		return nil, err
	}

	httpClient, err := httpClient.NewHttpClient(proxyPool, httpClient.DefaultConfig())
	if err != nil {
		return nil, err
	}

	return &debankConfig.DebankConfig{
		BaseURL:         cfg.BaseURL,
		Endpoints:       cfg.Endpoints,
		ContextDeadline: cfg.ContextDeadline,
		HttpClient:      httpClient,
	}, nil
}

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

func initHttpClient(debankSettings *appConfig.DebankSettings) (httpClient.HttpClientInterface, error) {
	cfg := httpConfig.DefaultConfig()
	cfg.UseProxyPool = debankSettings.UseProxyPool
	cfg.IsRotatingProxy = debankSettings.RotateProxy
	cfg.UseUTLS = true
	cfg.UTLSClientID = "Chrome_112"

	proxies, err := initProxies(debankSettings.ProxyFilePath)
	if err != nil {
		return nil, err
	}

	proxyPool, err := proxyPool.NewProxyPool(proxies)
	if err != nil {
		return nil, err
	}

	return httpFactory.NewHttpClient(cfg, proxyPool)
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
