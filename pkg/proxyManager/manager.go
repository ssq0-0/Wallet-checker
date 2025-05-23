package proxyManager

import (
	"fmt"
)

// ProxyManagerImpl implements the ProxyManager interface
type ProxyManagerImpl struct {
	proxies   []*ProxyConfig
	parser    ProxyParser
	validator ProxyValidator
	formatter ProxyFormatter
}

// NewProxyManager creates a new instance of ProxyManagerImpl
func NewProxyManager(parser ProxyParser, validator ProxyValidator, formatter ProxyFormatter) *ProxyManagerImpl {
	return &ProxyManagerImpl{
		proxies:   make([]*ProxyConfig, 0),
		parser:    parser,
		validator: validator,
		formatter: formatter,
	}
}

// ParseAndValidate implements ProxyManager interface
func (p *ProxyManagerImpl) ParseAndValidate(proxyStr string) (*ProxyConfig, error) {
	config, err := p.parser.Parse(proxyStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse proxy: %w", err)
	}

	if err := p.validator.Validate(config); err != nil {
		return nil, fmt.Errorf("proxy validation failed: %w", err)
	}

	return config, nil
}

// FormatProxy implements ProxyManager interface
func (p *ProxyManagerImpl) FormatProxy(config *ProxyConfig) string {
	return p.formatter.Format(config)
}

// AddProxy implements ProxyManager interface
func (p *ProxyManagerImpl) AddProxy(proxyStr string) error {
	config, err := p.ParseAndValidate(proxyStr)
	if err != nil {
		return err
	}

	p.proxies = append(p.proxies, config)
	return nil
}

// RemoveProxy implements ProxyManager interface
func (p *ProxyManagerImpl) RemoveProxy(proxyStr string) error {
	config, err := p.ParseAndValidate(proxyStr)
	if err != nil {
		return err
	}

	for i, proxy := range p.proxies {
		if proxy.Host == config.Host && proxy.Port == config.Port {
			p.proxies = append(p.proxies[:i], p.proxies[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("proxy not found")
}

// GetProxies implements ProxyManager interface
func (p *ProxyManagerImpl) GetProxies() []*ProxyConfig {
	return p.proxies
}
