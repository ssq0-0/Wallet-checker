package debankConfig

import (
	"chief-checker/internal/infrastructure/httpClient/httpInterfaces"
)

type DebankConfig struct {
	BaseURL         string
	Endpoints       map[string]string
	ContextDeadline int
	HttpClient      httpInterfaces.HttpClientInterface
}
