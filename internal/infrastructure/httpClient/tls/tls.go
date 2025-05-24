// Package tls provides advanced TLS client functionality for HTTP connections.
// It supports both standard Go TLS and uTLS (https://github.com/refraction-networking/utls)
// implementations for client fingerprint emulation.
//
// The package includes:
// - TLS and uTLS connection dialing with configurable parameters
// - Browser fingerprint emulation (Chrome, Firefox, Safari)
// - Transport configuration with randomized parameters
// - Various utilities for TLS configuration
//
// This package helps with bypassing anti-bot systems and emulating real browser
// TLS fingerprints for improved web scraping capabilities.
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

// BrowserSpec represents a browser TLS fingerprint specification.
// It contains all the necessary information to emulate a specific browser's TLS handshake.
type BrowserSpec struct {
	clientHelloID utls.ClientHelloID  // The uTLS client hello ID for a specific browser version
	extensions    []utls.TLSExtension // TLS extensions to include in the handshake
	cipherSuites  []uint16            // Cipher suites supported by the browser
}

// StandardTLSDialer implements the standard Go TLS dialer.
// It uses the native crypto/tls package for TLS connections.
type StandardTLSDialer struct {
	serverName string // Server name for TLS certificate validation
}

// NewStandardTLSDialer creates a new StandardTLSDialer with the specified server name.
//
// Parameters:
//   - serverName: The hostname to use for certificate validation
//
// Returns:
//   - A new StandardTLSDialer instance
func NewStandardTLSDialer(serverName string) *StandardTLSDialer {
	return &StandardTLSDialer{
		serverName: serverName,
	}
}

// DialTLSContext establishes a TLS connection using the standard Go TLS implementation.
// It implements the httpInterfaces.TLSDialer interface.
//
// Parameters:
//   - ctx: Context for the connection timeout and cancellation
//   - network: The network type ("tcp", "tcp4", "tcp6")
//   - addr: The address to connect to ("host:port")
//
// Returns:
//   - A TLS connection or an error if the connection fails
func (d *StandardTLSDialer) DialTLSContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, _, _ := net.SplitHostPort(addr)
	if host == "" {
		host = d.serverName
	}

	dialer := &net.Dialer{
		Timeout:   time.Duration(7+rand.Intn(6)) * time.Second, // 7-12 seconds
		KeepAlive: -1,
	}

	tcpConn, err := dialer.DialContext(ctx, network, addr)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: false,
		MinVersion:         tls.VersionTLS12,
		MaxVersion:         tls.VersionTLS13,
		CurvePreferences: []tls.CurveID{
			tls.X25519,
			tls.CurveP256,
			tls.CurveP384,
		},
	}

	return tls.Client(tcpConn, tlsConfig), nil
}

// UTLSDialer implements a TLS dialer using the uTLS library for browser fingerprint emulation.
// It allows connections that mimic specific browser TLS fingerprints.
type UTLSDialer struct {
	serverName string // Server name for TLS certificate validation
}

// NewUTLSDialer creates a new UTLSDialer with the specified server name.
//
// Parameters:
//   - serverName: The hostname to use for certificate validation
//
// Returns:
//   - A new UTLSDialer instance
func NewUTLSDialer(serverName string) *UTLSDialer {
	return &UTLSDialer{
		serverName: serverName,
	}
}

// DialTLSContext establishes a TLS connection using the uTLS library.
// It randomly selects a browser fingerprint from predefined browser specifications.
// Implements the httpInterfaces.TLSDialer interface.
//
// The method includes random delays and parameters to mimic real browser behavior.
//
// Parameters:
//   - ctx: Context for the connection timeout and cancellation
//   - network: The network type ("tcp", "tcp4", "tcp6")
//   - addr: The address to connect to ("host:port")
//
// Returns:
//   - A uTLS connection mimicking a browser or an error if the connection fails
func (d *UTLSDialer) DialTLSContext(ctx context.Context, network, addr string) (net.Conn, error) {
	seed := time.Now().UnixNano() + int64(rand.Intn(10000))
	localRand := rand.New(rand.NewSource(seed))

	time.Sleep(time.Duration(50+localRand.Intn(100)) * time.Millisecond)

	host, _, _ := net.SplitHostPort(addr)
	if host == "" {
		host = d.serverName
	}

	tcpConn, err := (&net.Dialer{
		Timeout:   time.Duration(8+localRand.Intn(4)) * time.Second, // 8-11 seconds
		KeepAlive: -1,                                               // Disable KeepAlive
		DualStack: true,
	}).DialContext(ctx, network, addr)
	if err != nil {
		return nil, err
	}

	uTlsConfig := &utls.Config{
		ServerName:         host,
		InsecureSkipVerify: false,
		MinVersion:         utls.VersionTLS12,
		MaxVersion:         utls.VersionTLS13,
		Renegotiation:      utls.RenegotiateOnceAsClient,
		Time: func() time.Time {
			return time.Now().Add(-time.Duration(localRand.Intn(60)) * time.Second)
		},
	}

	spec := browserSpecs[localRand.Intn(len(browserSpecs))]

	uTlsConn := utls.UClient(tcpConn, uTlsConfig, spec.clientHelloID)

	time.Sleep(time.Duration(10+localRand.Intn(90)) * time.Millisecond)

	if err := uTlsConn.Handshake(); err != nil {
		tcpConn.Close()
		return nil, err
	}

	return uTlsConn, nil
}

// RandomTLSConfigurator provides randomized TLS configuration.
// It implements the httpInterfaces.TLSConfigurator interface.
type RandomTLSConfigurator struct{}

// NewRandomTLSConfigurator creates a new RandomTLSConfigurator.
//
// Returns:
//   - A new RandomTLSConfigurator instance
func NewRandomTLSConfigurator() *RandomTLSConfigurator {
	return &RandomTLSConfigurator{}
}

// ConfigureTLS randomly configures TLS parameters like cipher suites and curve preferences.
// It implements the httpInterfaces.TLSConfigurator interface.
//
// Parameters:
//   - config: The TLS configuration to modify
func (c *RandomTLSConfigurator) ConfigureTLS(config *tls.Config) {
	seed := time.Now().UnixNano()
	r := rand.New(rand.NewSource(seed))

	config.MinVersion = tls.VersionTLS12
	config.MaxVersion = tls.VersionTLS13

	for i := range cipherSuites {
		j := r.Intn(len(cipherSuites))
		cipherSuites[i], cipherSuites[j] = cipherSuites[j], cipherSuites[i]
	}

	numCiphers := 3 + r.Intn(5)
	if numCiphers > len(cipherSuites) {
		numCiphers = len(cipherSuites)
	}
	config.CipherSuites = cipherSuites[:numCiphers]

	for i := range curves {
		j := r.Intn(len(curves))
		curves[i], curves[j] = curves[j], curves[i]
	}

	numCurves := 2 + r.Intn(3)
	if numCurves > len(curves) {
		numCurves = len(curves)
	}
	config.CurvePreferences = curves[:numCurves]
}

// TransportManager manages HTTP transport configuration with TLS/uTLS support.
// It provides functionality to create and configure HTTP transports with specific TLS parameters.
type TransportManager struct {
	dialer     httpInterfaces.TLSDialer       // TLS connection dialer (standard or uTLS)
	tlsConfig  httpInterfaces.TLSConfigurator // TLS configuration provider
	serverName string                         // Server name for TLS certificate validation
}

// NewTransportManager creates a new TransportManager with specified parameters.
//
// Parameters:
//   - serverName: The hostname to use for certificate validation
//   - useUTLS: Whether to use uTLS instead of standard TLS
//   - utlsClientID: The client ID string (currently unused, reserved for future use)
//
// Returns:
//   - A new TransportManager instance
func NewTransportManager(serverName string, useUTLS bool, utlsClientID string) *TransportManager {
	var dialer httpInterfaces.TLSDialer
	if useUTLS {
		dialer = NewUTLSDialer(serverName)
	} else {
		dialer = NewStandardTLSDialer(serverName)
	}

	return &TransportManager{
		dialer:     dialer,
		tlsConfig:  NewRandomTLSConfigurator(),
		serverName: serverName,
	}
}

// delayedDialer wraps a TLS dialer with artificial connection delays.
// This helps mimic real browser behavior and avoid detection.
type delayedDialer struct {
	baseDialer httpInterfaces.TLSDialer // The underlying TLS dialer
}

// DialTLSContext adds a random delay before establishing a TLS connection.
// Implements the httpInterfaces.TLSDialer interface.
//
// Parameters:
//   - ctx: Context for the connection timeout and cancellation
//   - network: The network type ("tcp", "tcp4", "tcp6")
//   - addr: The address to connect to ("host:port")
//
// Returns:
//   - A TLS connection or an error if the connection fails
func (d *delayedDialer) DialTLSContext(ctx context.Context, network, addr string) (net.Conn, error) {
	time.Sleep(time.Duration(100+rand.Intn(200)) * time.Millisecond)
	return d.baseDialer.DialTLSContext(ctx, network, addr)
}

// ConfigureTransport creates and configures an http.Transport with randomized parameters.
// The transport uses the TLS dialer and configurator from the TransportManager.
//
// Returns:
//   - A configured http.Transport instance ready for use with an HTTP client
func (t *TransportManager) ConfigureTransport() *http.Transport {
	seed := time.Now().UnixNano()
	r := rand.New(rand.NewSource(seed))

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			ServerName:         t.serverName,
			InsecureSkipVerify: false,
			MinVersion:         tls.VersionTLS12,
			MaxVersion:         tls.VersionTLS13,
			Renegotiation:      tls.RenegotiateOnceAsClient,
		},
		DisableKeepAlives:     true,
		MaxIdleConns:          r.Intn(10) + 10,                            // 10-19
		MaxIdleConnsPerHost:   1,                                          // Maximum 1 connection
		MaxConnsPerHost:       r.Intn(5) + 5,                              // 5-9
		IdleConnTimeout:       time.Duration(5+r.Intn(10)) * time.Second,  // 5-14 seconds
		TLSHandshakeTimeout:   time.Duration(10+r.Intn(10)) * time.Second, // 10-19 seconds
		ExpectContinueTimeout: time.Duration(1+r.Intn(2)) * time.Second,   // 1-2 seconds
		ForceAttemptHTTP2:     r.Intn(2) == 0,                             // Random HTTP2 choice
		WriteBufferSize:       (8 + r.Intn(16)) * 1024,                    // 8-23 KB
		ReadBufferSize:        (8 + r.Intn(16)) * 1024,                    // 8-23 KB
		ResponseHeaderTimeout: time.Duration(20+r.Intn(20)) * time.Second, // 20-39 seconds
	}

	delayedTLSDialer := &delayedDialer{baseDialer: t.dialer}
	transport.DialTLSContext = delayedTLSDialer.DialTLSContext

	t.tlsConfig.ConfigureTLS(transport.TLSClientConfig)

	return transport
}

// GetTLSConfig returns a basic TLS configuration with reasonable defaults.
//
// Returns:
//   - A new tls.Config instance with basic security settings
func GetTLSConfig() *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: true,
		MinVersion:         tls.VersionTLS12,
		MaxVersion:         tls.VersionTLS13,
	}
}

// GetUTLSConfig returns a basic uTLS configuration with reasonable defaults.
//
// Returns:
//   - A new utls.Config instance with basic security settings
func GetUTLSConfig() *utls.Config {
	return &utls.Config{
		InsecureSkipVerify: true,
		MinVersion:         utls.VersionTLS12,
		MaxVersion:         utls.VersionTLS13,
	}
}

// GetClientHelloID returns a random Chrome browser ClientHelloID for uTLS.
//
// Returns:
//   - A random Chrome browser ClientHelloID
func GetClientHelloID() utls.ClientHelloID {
	return chromeHelloIds[rand.Intn(len(chromeHelloIds))]
}

// GetFirefoxClientHelloID returns a random Firefox browser ClientHelloID for uTLS.
//
// Returns:
//   - A random Firefox browser ClientHelloID
func GetFirefoxClientHelloID() utls.ClientHelloID {
	return firefoxHelloIds[rand.Intn(len(firefoxHelloIds))]
}
