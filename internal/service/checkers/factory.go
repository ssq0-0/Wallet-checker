// Package checkers provides implementations for various balance checking services.
package checkers

import (
	"chief-checker/internal/config/serviceConfig/debankConfig"
	"chief-checker/internal/infrastructure/wasmClient"
	"chief-checker/internal/service/checkers/adapters"
	"chief-checker/pkg/errors"
)

// Factory is responsible for creating checker instances.
// It ensures proper initialization of all dependencies and configurations.
type Factory struct {
	config *debankConfig.DebankConfig
}

// NewFactory creates a new instance of Factory with the provided configuration.
// It validates the configuration before creating the factory.
func NewFactory(cfg *debankConfig.DebankConfig) *Factory {
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
	wasmClient, err := wasmClient.NewWasm()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create wasm client")
	}

	idGenerator := adapters.NewIDGenerator()
	paramGenerator := adapters.NewParamGenerator(idGenerator, wasmClient)
	cache := adapters.NewMemoryCache()

	baseApiClient := adapters.NewApiChecker(f.config.BaseURL, f.config.Endpoints, f.config.HttpClient, cache, paramGenerator, f.config.ContextDeadline)

	return &Debank{
		baseChecker: baseApiClient,
		cache:       cache,
		ctxDeadline: f.config.ContextDeadline,
	}, nil
}
