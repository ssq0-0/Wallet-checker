package httpInterfaces

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
)

// TLSDialer определяет интерфейс для создания TLS соединений
type TLSDialer interface {
	DialTLSContext(ctx context.Context, network, addr string) (net.Conn, error)
}

// TLSConfigurator определяет интерфейс для настройки TLS конфигурации
type TLSConfigurator interface {
	ConfigureTLS(*tls.Config)
}

// TransportConfigurator определяет интерфейс для настройки HTTP транспорта
type TransportConfigurator interface {
	ConfigureTransport(*http.Transport)
}
