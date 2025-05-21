package proxyManager

type ProxyManager interface {
	ParseProxy(proxy []string) ([]string, error)
}
