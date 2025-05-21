package proxyManager

type ProxyManagerImpl struct {
	proxies []string
}

func NewProxyManager(proxies []string) (*ProxyManagerImpl, error) {
	return &ProxyManagerImpl{proxies: proxies}, nil
}

func (p *ProxyManagerImpl) ParseProxy(proxy []string) ([]string, error) { return proxy, nil }
