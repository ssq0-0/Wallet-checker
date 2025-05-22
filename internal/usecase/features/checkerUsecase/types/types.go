// Package types defines the data structures used throughout the checker system.
// It includes both raw and aggregated data structures for blockchain account information.
package types

import "chief-checker/internal/service/checkers/checkerModels/debankModels"

// RawAccountData represents the raw data collected from blockchain services.
// It contains the total balance and detailed information about tokens across chains.
type RawAccountData struct {
	// TotalBalance represents the total USD value of all assets
	TotalBalance float64

	// ChainsInfo maps chain names to lists of token information on that chain
	ChainsInfo map[string][]*TokenChainInfo

	// ProjectsInfo contains information about DeFi projects and their assets
	ProjectsInfo []*debankModels.ProjectAssets
}

// TokenChainInfo represents detailed information about a token on a specific chain.
type TokenChainInfo struct {
	// Amount is the token quantity
	Amount float64

	// Chain is the blockchain network name
	Chain string

	// Contract is the token's contract address
	Contract string

	// Price is the current token price in USD
	Price float64

	// Symbol is the token's symbol (e.g., "ETH", "USDT")
	Symbol string

	// UsdValue is the total USD value of the token holding
	UsdValue float64
}

// AggregatedData represents processed and aggregated account information.
// It organizes token data by chain and includes project information.
type AggregatedData struct {
	// Address is the blockchain address being analyzed
	Address string

	// TotalBalance is the total USD value of all assets
	TotalBalance float64

	// ChainData maps chain names to lists of token information
	ChainData map[string][]*TokenInfo

	// ProjectData contains information about DeFi projects
	ProjectData []*ProjectInfo
}

// TokenInfo represents processed token information with value calculations.
type TokenInfo struct {
	// Symbol is the token's symbol (e.g., "ETH", "USDT")
	Symbol string

	// Amount is the token quantity
	Amount float64

	// UsdValue is the total USD value of the token holding
	UsdValue float64

	// Contract is the token's contract address
	Contract string

	// Chain is the blockchain network name
	Chain string

	// Price is the current token price in USD
	Price float64
}

// ProjectInfo represents information about a DeFi project and its assets.
type ProjectInfo struct {
	// Name is the project's name
	Name string

	// SiteUrl is the project's website URL
	SiteUrl string

	// Chain is the blockchain network the project is on
	Chain string

	// Tokens contains information about tokens in the project
	Tokens []*TokenInfo
}

// GlobalStats represents system-wide statistics about processed accounts.
type GlobalStats struct {
	// TotalAccounts is the number of accounts processed
	TotalAccounts int32

	// TokenStats maps token symbols to aggregated token information
	TokenStats map[string]*TokenInfo

	// TotalUsdValue is the total USD value across all accounts
	TotalUsdValue float64
}
