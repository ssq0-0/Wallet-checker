package serviceConfig

import "chief-checker/internal/infrastructure/httpClient/httpInterfaces"

type ApiCheckerConfig struct {
	BaseURL         string
	Endpoints       map[string]string
	ContextDeadline int
	HttpClient      httpInterfaces.HttpClientInterface
}
