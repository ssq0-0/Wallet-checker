package httpFactory

import (
	"chief-checker/internal/infrastructure/httpClient/httpConfig"
	"chief-checker/internal/infrastructure/httpClient/tls"
	"net/http"
)

type TransportFactory struct{}

func NewTransportFactory() *TransportFactory {
	return &TransportFactory{}
}

func (f *TransportFactory) CreateTransport(config *httpConfig.Config) http.RoundTripper {
	transportManager := tls.NewTransportManager(
		config.Servername,
		config.UseUTLS,
		config.UTLSClientID,
	)

	return transportManager.ConfigureTransport()
}
