package types

import "chief-checker/internal/service/checkers/checkerModels/debankModels"

type RawAccountData struct {
	TotalBalance float64

	ChainsInfo map[string][]*TokenChainInfo

	ProjectsInfo []*debankModels.ProjectAssets
}

type TokenChainInfo struct {
	Amount float64

	Chain string

	Contract string

	Price float64

	Symbol string

	UsdValue float64
}

type AggregatedData struct {
	Address string

	TotalBalance float64

	ChainData map[string][]*TokenInfo

	ProjectData []*ProjectInfo
}

type TokenInfo struct {
	Symbol string

	Amount float64

	UsdValue float64

	Contract string

	Chain string

	Price float64
}

type ProjectInfo struct {
	Name string

	SiteUrl string

	Chain string

	Tokens []*TokenInfo
}

type GlobalStats struct {
	TotalAccounts int32

	TokenStats map[string]*TokenInfo

	TotalUsdValue float64
}
