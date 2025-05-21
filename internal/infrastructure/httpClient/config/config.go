package config

import "time"

// Config holds the configuration for the HTTP client
type Config struct {
	// UseProxyPool indicates whether to use proxy pool
	UseProxyPool bool

	// IsRotatingProxy indicates whether to rotate proxies
	IsRotatingProxy bool

	// UseUTLS indicates whether to use uTLS
	UseUTLS bool

	// UTLSClientID specifies the uTLS client ID
	UTLSClientID string

	// MaxRetries specifies the maximum number of retries
	MaxRetries int

	// RetryDelay specifies the delay between retries
	RetryDelay time.Duration

	// Timeout specifies the client timeout
	Timeout time.Duration
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		UseProxyPool:    false,
		IsRotatingProxy: false,
		UseUTLS:         false,
		UTLSClientID:    "Chrome_112",
		MaxRetries:      3,
		RetryDelay:      time.Second * 2,
		Timeout:         time.Second * 30,
	}
}
