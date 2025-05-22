package tls

import (
	"chief-checker/internal/infrastructure/httpClient/httpInterfaces"
	"context"
	"crypto/tls"
	"math/rand"
	"net"
	"net/http"
	"time"

	utls "github.com/refraction-networking/utls"
)

// StandardTLSDialer реализует стандартный TLS диалер
type StandardTLSDialer struct {
	serverName string
}

// NewStandardTLSDialer создает новый стандартный TLS диалер
func NewStandardTLSDialer(serverName string) *StandardTLSDialer {
	return &StandardTLSDialer{
		serverName: serverName,
	}
}

// DialTLSContext реализует интерфейс TLSDialer
func (d *StandardTLSDialer) DialTLSContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, _, _ := net.SplitHostPort(addr)
	if host == "" {
		host = d.serverName
	}

	tcpConn, err := (&net.Dialer{Timeout: 10 * time.Second}).DialContext(ctx, network, addr)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: false,
	}

	return tls.Client(tcpConn, tlsConfig), nil
}

// UTLSDialer реализует uTLS диалер
type UTLSDialer struct {
	serverName string
	clientID   string
}

// NewUTLSDialer создает новый uTLS диалер
func NewUTLSDialer(serverName, clientID string) *UTLSDialer {
	return &UTLSDialer{
		serverName: serverName,
		clientID:   clientID,
	}
}

// DialTLSContext реализует интерфейс TLSDialer
func (d *UTLSDialer) DialTLSContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, _, _ := net.SplitHostPort(addr)
	if host == "" {
		host = d.serverName
	}

	tcpConn, err := (&net.Dialer{Timeout: 10 * time.Second}).DialContext(ctx, network, addr)
	if err != nil {
		return nil, err
	}

	uTlsConfig := &utls.Config{
		ServerName:         host,
		InsecureSkipVerify: false,
	}

	var uTlsConn *utls.UConn
	switch d.clientID {
	case "Chrome_112":
		uTlsConn = utls.UClient(tcpConn, uTlsConfig, utls.HelloChrome_Auto)
	default:
		uTlsConn = utls.UClient(tcpConn, uTlsConfig, utls.HelloChrome_Auto)
	}

	if err := uTlsConn.Handshake(); err != nil {
		return nil, err
	}

	return uTlsConn, nil
}

// RandomTLSConfigurator реализует случайную конфигурацию TLS
type RandomTLSConfigurator struct{}

// NewRandomTLSConfigurator создает новый конфигуратор случайного TLS
func NewRandomTLSConfigurator() *RandomTLSConfigurator {
	return &RandomTLSConfigurator{}
}

// ConfigureTLS реализует интерфейс TLSConfigurator
func (c *RandomTLSConfigurator) ConfigureTLS(config *tls.Config) {
	config.MinVersion = tls.VersionTLS12
	config.MaxVersion = tls.VersionTLS13

	cipherSuites := []uint16{
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range cipherSuites {
		j := r.Intn(len(cipherSuites))
		cipherSuites[i], cipherSuites[j] = cipherSuites[j], cipherSuites[i]
	}
	config.CipherSuites = cipherSuites[:3+r.Intn(3)]
}

// TransportManager управляет HTTP транспортом
type TransportManager struct {
	dialer     httpInterfaces.TLSDialer
	tlsConfig  httpInterfaces.TLSConfigurator
	serverName string
}

// NewTransportManager создает новый менеджер транспорта
func NewTransportManager(serverName string, useUTLS bool, utlsClientID string) *TransportManager {
	var dialer httpInterfaces.TLSDialer
	if useUTLS {
		dialer = NewUTLSDialer(serverName, utlsClientID)
	} else {
		dialer = NewStandardTLSDialer(serverName)
	}

	return &TransportManager{
		dialer:     dialer,
		tlsConfig:  NewRandomTLSConfigurator(),
		serverName: serverName,
	}
}

// ConfigureTransport создает и настраивает HTTP транспорт
func (t *TransportManager) ConfigureTransport() *http.Transport {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			ServerName:         t.serverName,
			InsecureSkipVerify: false,
		},
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		MaxConnsPerHost:     100,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 30 * time.Second,
		DisableKeepAlives:   true,
		ForceAttemptHTTP2:   false,
	}

	transport.DialTLSContext = t.dialer.DialTLSContext
	t.tlsConfig.ConfigureTLS(transport.TLSClientConfig)

	return transport
}
