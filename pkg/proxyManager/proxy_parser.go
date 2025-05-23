package proxyManager

import (
	"fmt"
	"net"
	"regexp"
	"strings"
)

// Supported proxy protocols
const (
	ProtocolHTTP  = "http"
	ProtocolHTTPS = "https"
	ProtocolSOCKS = "socks5"
)

// Error messages
const (
	ErrEmptyProxyString = "empty proxy string"
	ErrInvalidFormat    = "invalid proxy format"
	ErrInvalidHost      = "invalid host: %s"
	ErrInvalidPort      = "invalid port: %s"
	ErrInvalidProtocol  = "invalid protocol: %s"
)

// DefaultProxyPort is the default port used when no port is specified
const DefaultProxyPort = "8080"

// proxyRegexes defines regular expressions for various proxy formats
var proxyRegexes = map[string]*regexp.Regexp{
	// http://user:pass@host:port
	"http_auth_at": regexp.MustCompile(`^http://([^:@]+):([^:@]+)@([^:@]+):(\d+)$`),
	// http://user:pass:host:port
	"http_auth_colon": regexp.MustCompile(`^http://([^:@]+):([^:@]+):([^:@]+):(\d+)$`),
	// http://host:port@user:pass
	"http_host_at": regexp.MustCompile(`^http://([^:@]+):(\d+)@([^:@]+):([^:@]+)$`),
	// http://host:port:user:pass
	"http_host_colon": regexp.MustCompile(`^http://([^:@]+):(\d+):([^:@]+):([^:@]+)$`),

	// https://user:pass@host:port
	"https_auth_at": regexp.MustCompile(`^https://([^:@]+):([^:@]+)@([^:@]+):(\d+)$`),
	// https://user:pass:host:port
	"https_auth_colon": regexp.MustCompile(`^https://([^:@]+):([^:@]+):([^:@]+):(\d+)$`),
	// https://host:port@user:pass
	"https_host_at": regexp.MustCompile(`^https://([^:@]+):(\d+)@([^:@]+):([^:@]+)$`),
	// https://host:port:user:pass
	"https_host_colon": regexp.MustCompile(`^https://([^:@]+):(\d+):([^:@]+):([^:@]+)$`),

	// socks5://user:pass@host:port
	"socks5_auth_at": regexp.MustCompile(`^socks5://([^:@]+):([^:@]+)@([^:@]+):(\d+)$`),
	// socks5://user:pass:host:port
	"socks5_auth_colon": regexp.MustCompile(`^socks5://([^:@]+):([^:@]+):([^:@]+):(\d+)$`),
	// socks5://host:port@user:pass
	"socks5_host_at": regexp.MustCompile(`^socks5://([^:@]+):(\d+)@([^:@]+):([^:@]+)$`),
	// socks5://host:port:user:pass
	"socks5_host_colon": regexp.MustCompile(`^socks5://([^:@]+):(\d+):([^:@]+):([^:@]+)$`),

	// user:pass@host:port
	"auth_at": regexp.MustCompile(`^([^:@]+):([^:@]+)@([^:@]+):(\d+)$`),
	// user:pass:host:port
	"auth_colon": regexp.MustCompile(`^([^:@]+):([^:@]+):([^:@]+):(\d+)$`),
	// host:port@user:pass
	"host_at": regexp.MustCompile(`^([^:@]+):(\d+)@([^:@]+):([^:@]+)$`),
	// host:port:user:pass
	"host_colon": regexp.MustCompile(`^([^:@]+):(\d+):([^:@]+):([^:@]+)$`),
}

// DefaultProxyParser implements the ProxyParser interface
type DefaultProxyParser struct{}

// NewDefaultProxyParser creates a new instance of DefaultProxyParser
func NewDefaultProxyParser() *DefaultProxyParser {
	return &DefaultProxyParser{}
}

// Parse implements the ProxyParser interface
func (p *DefaultProxyParser) Parse(proxyStr string) (*ProxyConfig, error) {
	if proxyStr == "" {
		return nil, fmt.Errorf(ErrEmptyProxyString)
	}

	for pattern, re := range proxyRegexes {
		matches := re.FindStringSubmatch(proxyStr)
		if matches != nil {
			config := &ProxyConfig{}

			// Determine protocol from pattern
			switch {
			case strings.HasPrefix(pattern, "http_"):
				config.Protocol = ProtocolHTTP
			case strings.HasPrefix(pattern, "https_"):
				config.Protocol = ProtocolHTTPS
			case strings.HasPrefix(pattern, "socks5_"):
				config.Protocol = ProtocolSOCKS
			default:
				config.Protocol = ProtocolHTTP // Default to HTTP
			}

			// Determine field order based on format
			switch {
			case strings.Contains(pattern, "auth_at"), strings.Contains(pattern, "auth_colon"):
				config.Username = matches[1]
				config.Password = matches[2]
				config.Host = matches[3]
				config.Port = matches[4]
			case strings.Contains(pattern, "host_at"), strings.Contains(pattern, "host_colon"):
				config.Host = matches[1]
				config.Port = matches[2]
				config.Username = matches[3]
				config.Password = matches[4]
			}

			// Validate the config
			if err := validateConfig(config); err != nil {
				return nil, err
			}

			return config, nil
		}
	}

	return nil, fmt.Errorf(ErrInvalidFormat)
}

// validateConfig performs validation of the proxy configuration
func validateConfig(config *ProxyConfig) error {
	// Validate host
	if net.ParseIP(config.Host) == nil && !IsValidHostname(config.Host) {
		return fmt.Errorf(ErrInvalidHost, config.Host)
	}

	// Validate port
	if port, err := net.LookupPort("tcp", config.Port); err != nil || port < 1 || port > 65535 {
		return fmt.Errorf(ErrInvalidPort, config.Port)
	}

	// Validate protocol
	switch config.Protocol {
	case ProtocolHTTP, ProtocolHTTPS, ProtocolSOCKS:
		// Valid protocols
	default:
		return fmt.Errorf(ErrInvalidProtocol, config.Protocol)
	}

	return nil
}

// DefaultProxyValidator implements the ProxyValidator interface
type DefaultProxyValidator struct{}

// NewDefaultProxyValidator creates a new instance of DefaultProxyValidator
func NewDefaultProxyValidator() *DefaultProxyValidator {
	return &DefaultProxyValidator{}
}

// Validate implements the ProxyValidator interface
func (v *DefaultProxyValidator) Validate(config *ProxyConfig) error {
	// Validate host
	if net.ParseIP(config.Host) == nil && !IsValidHostname(config.Host) {
		return fmt.Errorf(ErrInvalidHost, config.Host)
	}

	// Validate port
	if port, err := net.LookupPort("tcp", config.Port); err != nil || port < 1 || port > 65535 {
		return fmt.Errorf(ErrInvalidPort, config.Port)
	}

	// Validate protocol
	switch config.Protocol {
	case ProtocolHTTP, ProtocolHTTPS, ProtocolSOCKS:
		// Valid protocols
	default:
		return fmt.Errorf(ErrInvalidProtocol, config.Protocol)
	}

	return nil
}

// DefaultProxyFormatter implements the ProxyFormatter interface
type DefaultProxyFormatter struct{}

// NewDefaultProxyFormatter creates a new instance of DefaultProxyFormatter
func NewDefaultProxyFormatter() *DefaultProxyFormatter {
	return &DefaultProxyFormatter{}
}

// Format implements the ProxyFormatter interface
func (f *DefaultProxyFormatter) Format(config *ProxyConfig) string {
	var builder strings.Builder

	// Add protocol
	if config.Protocol != "" {
		builder.WriteString(config.Protocol)
		builder.WriteString("://")
	}

	// Add credentials if present
	if config.Username != "" && config.Password != "" {
		builder.WriteString(config.Username)
		builder.WriteString(":")
		builder.WriteString(config.Password)
		builder.WriteString("@")
	}

	// Add host:port
	builder.WriteString(config.Host)
	builder.WriteString(":")
	builder.WriteString(config.Port)

	return builder.String()
}

// IsValidHostname checks if a string is a valid hostname
func IsValidHostname(hostname string) bool {
	if len(hostname) > 255 {
		return false
	}

	hostnameRegex := regexp.MustCompile(`^[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)
	return hostnameRegex.MatchString(hostname)
}
