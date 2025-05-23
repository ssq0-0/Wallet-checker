package proxyManager

// ProxyConfig represents a standardized proxy configuration
type ProxyConfig struct {
	Protocol string
	Host     string
	Port     string
	Username string
	Password string
}

// ProxyParser defines an interface for parsing proxy strings into ProxyConfig
type ProxyParser interface {
	// Parse converts a proxy string into a ProxyConfig
	// Returns error if the string format is invalid
	Parse(proxyStr string) (*ProxyConfig, error)
}

// ProxyValidator defines an interface for validating proxy configurations
type ProxyValidator interface {
	// Validate checks if the proxy configuration is valid
	// Returns error if validation fails
	Validate(config *ProxyConfig) error
}

// ProxyFormatter defines an interface for formatting proxy configurations
type ProxyFormatter interface {
	// Format converts a ProxyConfig into a standardized string representation
	Format(config *ProxyConfig) string
}

// ProxyManager defines the main interface for proxy management operations
type ProxyManager interface {
	// ParseAndValidate parses a proxy string and validates the configuration
	ParseAndValidate(proxyStr string) (*ProxyConfig, error)

	// FormatProxy formats a proxy configuration into a string
	FormatProxy(config *ProxyConfig) string

	// AddProxy adds a new proxy to the manager
	AddProxy(proxyStr string) error

	// RemoveProxy removes a proxy from the manager
	RemoveProxy(proxyStr string) error

	// GetProxies returns all managed proxies
	GetProxies() []*ProxyConfig
}
