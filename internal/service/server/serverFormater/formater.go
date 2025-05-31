// Package serverFormater provides data formatting functionality for the server package.
// It implements the Formater interface to transform internal data structures into API responses.
package serverFormater

import (
	"chief-checker/internal/service/server/serverInterface"
	"chief-checker/internal/service/server/serverTypes"
	"chief-checker/internal/usecase/features/checkerUsecase/types"
	"sort"
)

const (
	maxTopTokens = 10
)

// formater implements the serverInterface.Formater interface.
// It provides methods to format balance and address data for API responses.
type formater struct{}

// NewFormater creates a new instance of the formater.
// It returns an implementation of the serverInterface.Formater interface.
func NewFormater() serverInterface.Formater { return &formater{} }

// BalanceDataFormat formats the global statistics into a structured response.
// It processes token and chain statistics to create a formatted response with:
// - Global statistics (total accounts and total value)
// - Top 10 tokens by value
// - Chain statistics
func (f *formater) BalanceDataFormat(stats *types.GlobalStats) *serverTypes.TotalStats {
	if stats == nil {
		return &serverTypes.TotalStats{}
	}

	chainStats := f.aggregateChainStats(stats.TokenStats)
	tokenStats := f.aggregateTokenStats(stats.TokenStats)
	topTokens := f.getTopTokens(tokenStats)

	return &serverTypes.TotalStats{
		GlobalStats: serverTypes.GlobalStats{
			TotalAccounts: stats.TotalAccounts,
			TotalValue:    stats.TotalUsdValue,
		},
		TopTokens: topTokens,
		Chains:    chainStats,
	}
}

// aggregateChainStats aggregates token statistics by chain.
func (f *formater) aggregateChainStats(tokens map[string]*types.TokenInfo) []serverTypes.ChainStat {
	chainMap := make(map[string]float64)
	for _, token := range tokens {
		chainMap[token.Chain] += token.UsdValue
	}

	chains := make([]serverTypes.ChainStat, 0, len(chainMap))
	for name, value := range chainMap {
		chains = append(chains, serverTypes.ChainStat{
			Name:       name,
			TotalValue: value,
		})
	}

	return chains
}

// aggregateTokenStats aggregates token statistics by symbol.
func (f *formater) aggregateTokenStats(tokens map[string]*types.TokenInfo) map[string]float64 {
	tokenMap := make(map[string]float64)
	for _, token := range tokens {
		tokenMap[token.Symbol] += token.UsdValue
	}
	return tokenMap
}

// getTopTokens returns the top N tokens by value.
func (f *formater) getTopTokens(tokenMap map[string]float64) []serverTypes.TopToken {
	topTokens := make([]serverTypes.TopToken, 0, len(tokenMap))
	for symbol, value := range tokenMap {
		topTokens = append(topTokens, serverTypes.TopToken{
			Symbol: symbol,
			Value:  value,
		})
	}

	sort.Slice(topTokens, func(i, j int) bool {
		return topTokens[i].Value > topTokens[j].Value
	})

	if len(topTokens) > maxTopTokens {
		topTokens = topTokens[:maxTopTokens]
	}

	return topTokens
}

// AddressDataFormat formats the address data into a structured response.
// It processes individual address data to create a formatted response containing:
// - Address information
// - Total balance
// - Token count
// - Project count
func (f *formater) AddressDataFormat(data map[string]*types.AggregatedData) []serverTypes.AddressResponse {
	if data == nil {
		return []serverTypes.AddressResponse{}
	}

	addresses := make([]serverTypes.AddressResponse, 0, len(data))
	for _, addrData := range data {
		addresses = append(addresses, f.formatAddressData(addrData))
	}

	return addresses
}

// getTopTokensForAddress returns the top N tokens by value for a specific address.
func (f *formater) getTopTokensForAddress(chainData map[string][]*types.TokenInfo) []serverTypes.TopToken {
	tokenMap := make(map[string]float64)
	for _, tokens := range chainData {
		for _, token := range tokens {
			tokenMap[token.Symbol] += token.UsdValue
		}
	}

	topTokens := make([]serverTypes.TopToken, 0, len(tokenMap))
	for symbol, value := range tokenMap {
		topTokens = append(topTokens, serverTypes.TopToken{
			Symbol: symbol,
			Value:  value,
		})
	}

	sort.Slice(topTokens, func(i, j int) bool {
		return topTokens[i].Value > topTokens[j].Value
	})

	if len(topTokens) > 5 {
		topTokens = topTokens[:5]
	}

	return topTokens
}

// getTopProjectsForAddress returns the top N projects by value for a specific address.
func (f *formater) getTopProjectsForAddress(projectData []*types.ProjectInfo) []serverTypes.TopProject {
	topProjects := make([]serverTypes.TopProject, 0, len(projectData))
	for _, project := range projectData {
		totalValue := 0.0
		for _, token := range project.Tokens {
			totalValue += token.UsdValue
		}
		topProjects = append(topProjects, serverTypes.TopProject{
			Name:  project.Name,
			Value: totalValue,
		})
	}

	sort.Slice(topProjects, func(i, j int) bool {
		return topProjects[i].Value > topProjects[j].Value
	})

	if len(topProjects) > 5 {
		topProjects = topProjects[:5]
	}

	return topProjects
}

// formatAddressData formats a single address data into a response.
func (f *formater) formatAddressData(data *types.AggregatedData) serverTypes.AddressResponse {
	tokenCount := f.calculateTokenCount(data.ChainData)
	topTokens := f.getTopTokensForAddress(data.ChainData)
	topProjects := f.getTopProjectsForAddress(data.ProjectData)

	return serverTypes.AddressResponse{
		Address:      data.Address,
		TotalBalance: data.TotalBalance,
		TokenCount:   tokenCount,
		ProjectCount: len(data.ProjectData),
		TopTokens:    topTokens,
		TopProjects:  topProjects,
	}
}

// calculateTokenCount calculates the total number of tokens across all chains.
func (f *formater) calculateTokenCount(chainData map[string][]*types.TokenInfo) int {
	tokenCount := 0
	for _, tokens := range chainData {
		tokenCount += len(tokens)
	}
	return tokenCount
}
