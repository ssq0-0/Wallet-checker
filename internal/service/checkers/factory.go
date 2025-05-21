package checkers

import (
	"chief-checker/internal/config/serviceConfig/debankConfig"
	"chief-checker/internal/infrastructure/wasmClient"
	"chief-checker/internal/service/checkers/adapters"
	"chief-checker/pkg/errors"
)

type Factory struct {
	config *debankConfig.DebankConfig
}

func NewFactory(cfg *debankConfig.DebankConfig) *Factory {
	return &Factory{
		config: cfg,
	}
}

func (f *Factory) CreateDebank() (*Debank, error) {
	wasmClient, err := wasmClient.NewWasm()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create wasm client")
	}

	idGenerator := adapters.NewIDGenerator()
	paramGenerator := adapters.NewParamGenerator(idGenerator, wasmClient)
	cache := adapters.NewMemoryCache()

	baseApiClient := adapters.NewApiChecker(f.config.Endpoints, f.config.HttpClient, cache, paramGenerator, f.config.ContextDeadline)

	return &Debank{
		baseChecker: baseApiClient,
		cache:       cache,
		ctxDeadline: f.config.ContextDeadline,
	}, nil
}
