package decorator

import (
	"chief-checker/internal/infrastructure/httpClient"
	"chief-checker/internal/infrastructure/proxyPool"
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	neturl "net/url"
	"time"
)

type proxyDecorator struct {
	client    httpClient.HttpClientInterface
	proxyPool proxyPool.ProxyPool
}

// NewProxyDecorator creates a new proxy decorator
func NewProxyDecorator(client httpClient.HttpClientInterface, proxyPool proxyPool.ProxyPool) httpClient.HttpClientInterface {
	return &proxyDecorator{
		client:    client,
		proxyPool: proxyPool,
	}
}

func (d *proxyDecorator) RequestWithRetry(ctx context.Context, url, method string, reqBody, respBody interface{}, headers map[string]string) error {
	proxy := d.proxyPool.GetFreeProxy()
	if proxy == "" {
		return fmt.Errorf("no free proxies available")
	}

	proxyURL, err := neturl.Parse(proxy)
	if err != nil {
		return fmt.Errorf("failed to parse proxy URL: %w", err)
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
			NextProtos:         []string{"http/1.1"},
		},
		DisableKeepAlives: true,
		ForceAttemptHTTP2: false,
		TLSNextProto:      make(map[string]func(string, *tls.Conn) http.RoundTripper),
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   d.client.GetTimeout(),
	}

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		d.proxyPool.BlockProxy(proxy, time.Minute*5)
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		d.proxyPool.BlockProxy(proxy, time.Minute*5)
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	d.proxyPool.UnblockProxy(proxy)
	return nil
}

func (d *proxyDecorator) SimpleRequest(ctx context.Context, url, method string, reqBody, respBody interface{}, headers map[string]string) error {
	return d.RequestWithRetry(ctx, url, method, reqBody, respBody, headers)
}

func (d *proxyDecorator) GetTimeout() time.Duration {
	return d.client.GetTimeout()
}
