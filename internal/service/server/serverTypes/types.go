package serverTypes

type AddressResponse struct {
	Address      string       `json:"address"`
	TotalBalance float64      `json:"totalBalance"`
	TokenCount   int          `json:"tokenCount"`
	ProjectCount int          `json:"projectCount"`
	TopTokens    []TopToken   `json:"topTokens"`
	TopProjects  []TopProject `json:"topProjects"`
}

type TotalStats struct {
	GlobalStats GlobalStats `json:"globalStats"`
	TopTokens   []TopToken  `json:"topTokens"`
	Chains      []ChainStat `json:"chains"`
}

type GlobalStats struct {
	TotalAccounts int32   `json:"totalAccounts"`
	TotalValue    float64 `json:"totalUSDValue"`
}

type TopToken struct {
	Symbol string  `json:"symbol"`
	Value  float64 `json:"value"`
}

type ChainStat struct {
	Name       string  `json:"name"`
	TotalValue float64 `json:"totalValue"`
}

type TopProject struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}
