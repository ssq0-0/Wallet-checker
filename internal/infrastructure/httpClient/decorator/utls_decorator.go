package decorator

import (
	"bytes"
	"chief-checker/internal/infrastructure/httpClient"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	utls "github.com/refraction-networking/utls"
)

type utlsDecorator struct {
	client     httpClient.HttpClientInterface
	clientID   string
	httpClient *http.Client
}

// NewUTLSDecorator creates a new uTLS decorator
func NewUTLSDecorator(client httpClient.HttpClientInterface, clientID string) httpClient.HttpClientInterface {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
			NextProtos:         []string{"http/1.1"},
		},
		DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, _, _ := net.SplitHostPort(addr)
			if host == "" {
				host = addr
			}

			tcpConn, err := (&net.Dialer{Timeout: 10 * time.Second}).DialContext(ctx, network, addr)
			if err != nil {
				return nil, err
			}

			uTlsConfig := &utls.Config{
				ServerName:         host,
				InsecureSkipVerify: false,
				NextProtos:         []string{"http/1.1"},
			}

			var uTlsConn *utls.UConn
			switch clientID {
			case "Chrome_112":
				uTlsConn = utls.UClient(tcpConn, uTlsConfig, utls.HelloChrome_Auto)
			default:
				uTlsConn = utls.UClient(tcpConn, uTlsConfig, utls.HelloChrome_Auto)
			}

			if err := uTlsConn.Handshake(); err != nil {
				return nil, err
			}

			return uTlsConn, nil
		},
		DisableKeepAlives:   true,
		ForceAttemptHTTP2:   false,
		TLSNextProto:        make(map[string]func(string, *tls.Conn) http.RoundTripper),
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		MaxConnsPerHost:     100,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	return &utlsDecorator{
		client:   client,
		clientID: clientID,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   client.GetTimeout(),
		},
	}
}

func (d *utlsDecorator) RequestWithRetry(ctx context.Context, url, method string, reqBody, respBody interface{}, headers map[string]string) error {
	return d.SimpleRequest(ctx, url, method, reqBody, respBody, headers)
}

func (d *utlsDecorator) SimpleRequest(ctx context.Context, url, method string, reqBody, respBody interface{}, headers map[string]string) error {
	var bodyReader io.Reader
	if reqBody != nil {
		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := d.httpClient.Do(req)
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

func (d *utlsDecorator) GetTimeout() time.Duration {
	return d.client.GetTimeout()
}
