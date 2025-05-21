package interfaces

import "chief-checker/internal/usecase/features/checkerUsecase/types"

type DataCollector interface {
	CollectData(address string) (*types.RawAccountData, error)
}

type DataAggregator interface {
	AggregateAccountData(address string, data *types.RawAccountData) (*types.AggregatedData, error)
	GetGlobalStats() *types.GlobalStats
	SetMinUsdAmount(amount float64)
}

type TokenCache interface {
	Update(symbol string, amount, usdValue float64)
	Get(symbol string) *types.TokenInfo
	GetAll() map[string]*types.TokenInfo
}

type Formatter interface {
	FormatAccountData(data *types.AggregatedData) ([]string, error)
	FormatGlobalStats(stats *types.GlobalStats) ([]string, error)
}
