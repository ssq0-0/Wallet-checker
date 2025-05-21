package debankConfig

import "chief-checker/internal/infrastructure/httpClient"

type DebankConfig struct {
	BaseURL         string
	Endpoints       map[string]string
	ContextDeadline int
	HttpClient      httpClient.HttpClientInterface
}
