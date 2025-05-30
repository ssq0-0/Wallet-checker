// Package interfaces defines the contracts for various components
// of the checker system.
package interfaces

import "chief-checker/internal/usecase/features/checkerUsecase/types"

// DataCollector defines the interface for collecting blockchain data.
// It is responsible for gathering raw account data from blockchain services.
type DataCollector interface {
	// CollectData retrieves data for a specific address.
	// Returns raw account data or an error if collection fails.
	CollectData(address string) (*types.RawAccountData, error)
}

// DataAggregator defines the interface for processing and aggregating collected data.
// It maintains global statistics and applies filtering based on minimum USD amounts.
type DataAggregator interface {
	// AggregateAccountData processes raw data for an address and returns aggregated results.
	// Returns nil data if the account doesn't meet minimum value requirements.
	AggregateAccountData(address string, data *types.RawAccountData) (*types.AggregatedData, error)

	// GetGlobalStats returns the current global statistics.
	GetGlobalStats() *types.GlobalStats
	GetAllStats() map[string]*types.AggregatedData
	// SetMinUsdAmount updates the minimum USD amount filter.
	SetMinUsdAmount(amount float64)
}

// TokenCache defines the interface for caching token information.
// It provides fast access to token data and maintains token statistics.
type TokenCache interface {
	// Update adds or updates token information in the cache.
	Update(token *types.TokenInfo)

	// Get retrieves information for a specific token.
	// Returns nil if the token is not in the cache.
	Get(symbol string) *types.TokenInfo

	// GetAll returns all cached token information.
	GetAll() map[string]*types.TokenInfo
}

// AccountStatCache defines the interface for caching account statistics.
// It provides fast access to aggregated account data and maintains a collection
// of all processed account statistics.
type AccountStatCache interface {
	// Update adds or updates aggregated data for a specific address in the cache.
	// Returns an error if the update operation fails.
	Update(address string, data *types.AggregatedData) error

	// GetAllStats returns a map of all cached account statistics,
	// where the key is the account address and the value is the aggregated data.
	GetAllStats() map[string]*types.AggregatedData
}

// Formatter defines the interface for formatting data into human-readable output.
// It handles both individual account data and global statistics.
type Formatter interface {
	// FormatAccountData formats aggregated account data into strings.
	// Returns formatted lines or an error if formatting fails.
	FormatAccountData(data *types.AggregatedData) ([]string, error)

	// FormatGlobalStats formats global statistics into strings.
	// Returns formatted lines or an error if formatting fails.
	FormatGlobalStats(stats *types.GlobalStats) ([]string, error)
}
