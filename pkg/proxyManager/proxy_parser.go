package proxyManager

import (
	"fmt"
	"net"
	"strings"
)

// Supported protocols
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

// DefaultProxyParser implements the proxy string parsing functionality
type DefaultProxyParser struct{}

// NewDefaultProxyParser creates a new instance of DefaultProxyParser
func NewDefaultProxyParser() *DefaultProxyParser {
	return &DefaultProxyParser{}
}

// Parse attempts to parse a proxy string into a ProxyConfig structure.
// It supports various formats including:
// - http://user:pass@host:port
// - https://user:pass:host:port
// - socks5://host:port@user:pass
// - host:port:user:pass
// - user:pass:host:port
// - host:port@user:pass
// - user:pass@host:port
// - host:port:user:pass
// - user:pass:host:port
// - host:port@user:pass
// - user:pass@host:port
// - host:port:user:pass
func (p *DefaultProxyParser) Parse(proxyStr string) (*ProxyConfig, error) {
	if proxyStr == "" {
		return nil, fmt.Errorf(ErrEmptyProxyString)
	}

	var protocol, username, password, host, port string

	protocol = ProtocolHTTP
	for _, prefix := range []string{"https://", "http://", "socks5://"} {
		if strings.HasPrefix(proxyStr, prefix) {
			protocol = strings.TrimSuffix(prefix, "://")
			proxyStr = strings.TrimPrefix(proxyStr, prefix)
			break
		}
	}

	if strings.Contains(proxyStr, "@") {
		parts := strings.Split(proxyStr, "@")
		if len(parts) != 2 {
			return nil, fmt.Errorf(ErrInvalidFormat)
		}

		authPart := parts[0]
		hostPart := parts[1]

		authParts := strings.Split(authPart, ":")
		if len(authParts) != 2 {
			return nil, fmt.Errorf(ErrInvalidFormat)
		}

		hostParts := strings.Split(hostPart, ":")
		if len(hostParts) != 2 {
			return nil, fmt.Errorf(ErrInvalidFormat)
		}

		if isIPOrDomain(authParts[0]) {
			host = authParts[0]
			port = authParts[1]
			username = hostParts[0]
			password = hostParts[1]
		} else {
			username = authParts[0]
			password = authParts[1]
			host = hostParts[0]
			port = hostParts[1]
		}
	} else {
		parts := strings.Split(proxyStr, ":")
		if len(parts) != 4 {
			return nil, fmt.Errorf(ErrInvalidFormat)
		}

		if isIPOrDomain(parts[0]) {
			host = parts[0]
			port = parts[1]
			username = parts[2]
			password = parts[3]
		} else {
			username = parts[0]
			password = parts[1]
			host = parts[2]
			port = parts[3]
		}
	}

	config := &ProxyConfig{
		Protocol: protocol,
		Username: username,
		Password: password,
		Host:     host,
		Port:     port,
	}

	validator := NewDefaultProxyValidator()
	if err := validator.Validate(config); err != nil {
		return nil, err
	}

	return config, nil
}

// DefaultProxyValidator implements proxy configuration validation
type DefaultProxyValidator struct{}

// NewDefaultProxyValidator creates a new instance of DefaultProxyValidator
func NewDefaultProxyValidator() *DefaultProxyValidator {
	return &DefaultProxyValidator{}
}

// Validate checks if the proxy configuration is valid
func (v *DefaultProxyValidator) Validate(config *ProxyConfig) error {
	// Check protocol
	switch config.Protocol {
	case ProtocolHTTP, ProtocolHTTPS, ProtocolSOCKS:
		// Valid protocols
	default:
		return fmt.Errorf(ErrInvalidProtocol, config.Protocol)
	}

	// Check host
	if !isValidHostname(config.Host) {
		return fmt.Errorf(ErrInvalidHost, config.Host)
	}

	// Check port
	if portNum, err := net.LookupPort("tcp", config.Port); err != nil || portNum < 1 || portNum > 65535 {
		return fmt.Errorf(ErrInvalidPort, config.Port)
	}

	return nil
}

// DefaultProxyFormatter implements proxy configuration formatting
type DefaultProxyFormatter struct{}

// NewDefaultProxyFormatter creates a new instance of DefaultProxyFormatter
func NewDefaultProxyFormatter() *DefaultProxyFormatter {
	return &DefaultProxyFormatter{}
}

// Format converts a ProxyConfig into a proxy string
func (f *DefaultProxyFormatter) Format(config *ProxyConfig) string {
	var builder strings.Builder

	if config.Protocol != "" {
		builder.WriteString(config.Protocol)
		builder.WriteString("://")
	}

	if config.Username != "" && config.Password != "" {
		builder.WriteString(config.Username)
		builder.WriteString(":")
		builder.WriteString(config.Password)
		builder.WriteString("@")
	}

	builder.WriteString(config.Host)
	builder.WriteString(":")
	builder.WriteString(config.Port)

	return builder.String()
}

// Helper functions
func isIPOrDomain(s string) bool {
	if net.ParseIP(s) != nil {
		return true
	}

	return strings.Contains(s, ".")
}

func isValidHostname(hostname string) bool {
	if net.ParseIP(hostname) != nil {
		return true
	}

	if strings.Contains(hostname, ".") {
		parts := strings.Split(hostname, ".")
		for _, part := range parts {
			if len(part) == 0 || len(part) > 63 {
				return false
			}
			if !isValidDomainPart(part) {
				return false
			}
		}
		return true
	}

	return false
}

func isValidDomainPart(s string) bool {
	if strings.HasPrefix(s, "-") || strings.HasSuffix(s, "-") {
		return false
	}
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_') {
			return false
		}
	}
	return true
}
