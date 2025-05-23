package proxyManager

import (
	"testing"
)

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

func TestProxyManager_Integration(t *testing.T) {
	parser := NewDefaultProxyParser()
	validator := NewDefaultProxyValidator()
	formatter := NewDefaultProxyFormatter()
	manager := NewProxyManager(parser, validator, formatter)

	// Test adding a proxy
	err := manager.AddProxy("http://user:pass@127.0.0.1:8080")
	if err != nil {
		t.Errorf("AddProxy() error = %v", err)
	}

	// Test getting proxies
	proxies := manager.GetProxies()
	if len(proxies) != 1 {
		t.Errorf("GetProxies() returned %d proxies, want 1", len(proxies))
	}

	// Test removing a proxy
	err = manager.RemoveProxy("http://user:pass@127.0.0.1:8080")
	if err != nil {
		t.Errorf("RemoveProxy() error = %v", err)
	}

	// Verify proxy was removed
	proxies = manager.GetProxies()
	if len(proxies) != 0 {
		t.Errorf("GetProxies() returned %d proxies, want 0", len(proxies))
	}
}

func TestIsValidHostname(t *testing.T) {
	tests := []struct {
		name     string
		hostname string
		want     bool
	}{
		{
			name:     "valid hostname",
			hostname: "example.com",
			want:     true,
		},
		{
			name:     "valid subdomain",
			hostname: "sub.example.com",
			want:     true,
		},
		{
			name:     "valid with numbers",
			hostname: "example123.com",
			want:     true,
		},
		{
			name:     "valid with hyphens",
			hostname: "my-example.com",
			want:     true,
		},
		{
			name:     "invalid double dots",
			hostname: "example..com",
			want:     false,
		},
		{
			name:     "invalid special chars",
			hostname: "example@.com",
			want:     false,
		},
		{
			name:     "invalid starts with hyphen",
			hostname: "-example.com",
			want:     false,
		},
		{
			name:     "invalid ends with hyphen",
			hostname: "example.com-",
			want:     false,
		},
		{
			name:     "too long",
			hostname: string(make([]byte, 256)),
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidHostname(tt.hostname); got != tt.want {
				t.Errorf("IsValidHostname() = %v, want %v", got, tt.want)
			}
		})
	}
}
