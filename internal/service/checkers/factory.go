// Package checkers provides implementations for various balance checking services.
package checkers

import (
	"chief-checker/internal/config/serviceConfig"
	"chief-checker/internal/service/checkers/adapters"
)

// Factory is responsible for creating checker instances.
// It ensures proper initialization of all dependencies and configurations.
type Factory struct {
	config *serviceConfig.ApiCheckerConfig
}

// NewFactory creates a new instance of Factory with the provided configuration.
// It validates the configuration before creating the factory.
func NewFactory(cfg *serviceConfig.ApiCheckerConfig) *Factory {
	return &Factory{
		config: cfg,
	}
}

// CreateDebank creates a new instance of Debank checker.
// It initializes all required dependencies:
// - WASM client for signature generation
// - ID generator for request identification
// - Parameter generator for request parameters
// - Memory cache for optimization
//
// Returns an error if any dependency initialization fails.
func (f *Factory) CreateDebank() (*Debank, error) {
	idGenerator := adapters.NewIDGenerator()
	paramGenerator := adapters.NewParamGeneratorImpl(idGenerator)
	cache := adapters.NewMemoryCache()

	baseApiClient := adapters.NewApiChecker(f.config.BaseURL, f.config.Endpoints, f.config.HttpClient, cache, paramGenerator, f.config.ContextDeadline)

	return &Debank{
		baseChecker: baseApiClient,
		cache:       cache,
		ctxDeadline: f.config.ContextDeadline,
	}, nil
}

func (f *Factory) CreateRabby() (*Rabby, error) {
	idGenerator := adapters.NewIDGenerator()
	paramGenerator := adapters.NewParamGeneratorImpl(idGenerator)
	cache := adapters.NewMemoryCache()

	baseApiClient := adapters.NewApiChecker(f.config.BaseURL, f.config.Endpoints, f.config.HttpClient, cache, paramGenerator, f.config.ContextDeadline)

	return &Rabby{
		baseChecker: baseApiClient,
		cache:       cache,
		ctxDeadline: f.config.ContextDeadline,
		httpClient:  f.config.HttpClient,
		endpoints:   f.config.Endpoints,
		baseUrl:     f.config.BaseURL,
	}, nil
}
