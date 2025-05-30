package proxyManager

import (
	"testing"
)

// TestDefaultProxyParser_Parse tests the proxy string parsing functionality
func TestDefaultProxyParser_Parse(t *testing.T) {
	parser := NewDefaultProxyParser()

	tests := []struct {
		name     string
		input    string
		expected *ProxyConfig
		wantErr  bool
	}{
		{
			name:  "http with user:pass@host:port",
			input: "http://user:pass@127.0.0.1:8080",
			expected: &ProxyConfig{
				Protocol: ProtocolHTTP,
				Username: "user",
				Password: "pass",
				Host:     "127.0.0.1",
				Port:     "8080",
			},
		},
		{
			name:  "https with user:pass:host:port",
			input: "https://user:pass:127.0.0.1:8080",
			expected: &ProxyConfig{
				Protocol: ProtocolHTTPS,
				Username: "user",
				Password: "pass",
				Host:     "127.0.0.1",
				Port:     "8080",
			},
		},
		{
			name:  "socks5 with host:port@user:pass",
			input: "socks5://127.0.0.1:8080@user:pass",
			expected: &ProxyConfig{
				Protocol: ProtocolSOCKS,
				Username: "user",
				Password: "pass",
				Host:     "127.0.0.1",
				Port:     "8080",
			},
		},
		{
			name:  "no protocol with host:port:user:pass",
			input: "127.0.0.1:8080:user:pass",
			expected: &ProxyConfig{
				Protocol: ProtocolHTTP,
				Username: "user",
				Password: "pass",
				Host:     "127.0.0.1",
				Port:     "8080",
			},
		},
		{
			name:  "valid hostname",
			input: "http://user:pass@example.com:8080",
			expected: &ProxyConfig{
				Protocol: ProtocolHTTP,
				Username: "user",
				Password: "pass",
				Host:     "example.com",
				Port:     "8080",
			},
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid format",
			input:   "invalid:format",
			wantErr: true,
		},
		{
			name:    "invalid host",
			input:   "http://user:pass@invalid..host:8080",
			wantErr: true,
		},
		{
			name:    "invalid port",
			input:   "http://user:pass@127.0.0.1:99999",
			wantErr: true,
		},
		{
			name:    "invalid protocol",
			input:   "ftp://user:pass@127.0.0.1:8080",
			wantErr: true,
		},
		{
			name:    "hostname too long",
			input:   "http://user:pass@" + string(make([]byte, 256)) + ":8080",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if got.Protocol != tt.expected.Protocol {
				t.Errorf("Protocol = %v, want %v", got.Protocol, tt.expected.Protocol)
			}
			if got.Username != tt.expected.Username {
				t.Errorf("Username = %v, want %v", got.Username, tt.expected.Username)
			}
			if got.Password != tt.expected.Password {
				t.Errorf("Password = %v, want %v", got.Password, tt.expected.Password)
			}
			if got.Host != tt.expected.Host {
				t.Errorf("Host = %v, want %v", got.Host, tt.expected.Host)
			}
			if got.Port != tt.expected.Port {
				t.Errorf("Port = %v, want %v", got.Port, tt.expected.Port)
			}
		})
	}
}

// TestDefaultProxyValidator_Validate tests the proxy configuration validation
func TestDefaultProxyValidator_Validate(t *testing.T) {
	validator := NewDefaultProxyValidator()

	tests := []struct {
		name    string
		config  *ProxyConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: &ProxyConfig{
				Protocol: ProtocolHTTP,
				Host:     "127.0.0.1",
				Port:     "8080",
			},
			wantErr: false,
		},
		{
			name: "invalid host",
			config: &ProxyConfig{
				Protocol: ProtocolHTTP,
				Host:     "invalid..host",
				Port:     "8080",
			},
			wantErr: true,
		},
		{
			name: "invalid port",
			config: &ProxyConfig{
				Protocol: ProtocolHTTP,
				Host:     "127.0.0.1",
				Port:     "99999",
			},
			wantErr: true,
		},
		{
			name: "invalid protocol",
			config: &ProxyConfig{
				Protocol: "ftp",
				Host:     "127.0.0.1",
				Port:     "8080",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestDefaultProxyFormatter_Format tests the proxy configuration formatting
func TestDefaultProxyFormatter_Format(t *testing.T) {
	formatter := NewDefaultProxyFormatter()

	tests := []struct {
		name     string
		config   *ProxyConfig
		expected string
	}{
		{
			name: "full config with protocol",
			config: &ProxyConfig{
				Protocol: ProtocolHTTP,
				Username: "user",
				Password: "pass",
				Host:     "127.0.0.1",
				Port:     "8080",
			},
			expected: "http://user:pass@127.0.0.1:8080",
		},
		{
			name: "without protocol",
			config: &ProxyConfig{
				Username: "user",
				Password: "pass",
				Host:     "127.0.0.1",
				Port:     "8080",
			},
			expected: "user:pass@127.0.0.1:8080",
		},
		{
			name: "without credentials",
			config: &ProxyConfig{
				Protocol: ProtocolHTTP,
				Host:     "127.0.0.1",
				Port:     "8080",
			},
			expected: "http://127.0.0.1:8080",
		},
		{
			name: "socks5 protocol",
			config: &ProxyConfig{
				Protocol: ProtocolSOCKS,
				Username: "user",
				Password: "pass",
				Host:     "127.0.0.1",
				Port:     "8080",
			},
			expected: "socks5://user:pass@127.0.0.1:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatter.Format(tt.config); got != tt.expected {
				t.Errorf("Format() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestProxyManager_Integration tests the integration of parser, validator and formatter
func TestProxyManager_Integration(t *testing.T) {
	parser := NewDefaultProxyParser()
	validator := NewDefaultProxyValidator()
	formatter := NewDefaultProxyFormatter()

	testCases := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "valid http proxy",
			input:    "http://user:pass@127.0.0.1:8080",
			expected: "http://user:pass@127.0.0.1:8080",
			wantErr:  false,
		},
		{
			name:     "valid https proxy",
			input:    "https://user:pass@127.0.0.1:8080",
			expected: "https://user:pass@127.0.0.1:8080",
			wantErr:  false,
		},
		{
			name:     "valid socks5 proxy",
			input:    "socks5://user:pass@127.0.0.1:8080",
			expected: "socks5://user:pass@127.0.0.1:8080",
			wantErr:  false,
		},
		{
			name:    "invalid proxy",
			input:   "invalid:format",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse
			config, err := parser.Parse(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if err != nil {
				return
			}

			// Validate
			if err := validator.Validate(config); err != nil {
				t.Errorf("Validate() error = %v", err)
				return
			}

			// Format
			got := formatter.Format(config)
			if got != tc.expected {
				t.Errorf("Format() = %v, want %v", got, tc.expected)
			}
		})
	}
}
