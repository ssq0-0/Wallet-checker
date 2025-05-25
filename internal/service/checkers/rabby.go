package checkers

import (
	"chief-checker/internal/config/serviceConfig"
	"chief-checker/internal/infrastructure/httpClient/httpInterfaces"
	"chief-checker/internal/service/checkers/checkerModels/commonModels"
	"chief-checker/internal/service/checkers/checkerModels/rabbyModels"
	"chief-checker/internal/service/checkers/port"
)

// Rabby implements the API client for interacting with Rabby wallet API.
// It provides methods to fetch wallet balances, used chains, token balances, and project assets.
type Rabby struct {
	baseChecker port.ApiClient
	httpClient  httpInterfaces.HttpClientInterface
	cache       port.Cache
	endpoints   map[string]string
	baseUrl     string
	ctxDeadline int
}

// NewRabby creates a new instance of Rabby client with the provided configuration.
func NewRabby(cfg *serviceConfig.ApiCheckerConfig) (*Rabby, error) {
	factory := NewFactory(cfg)
	return factory.CreateRabby()
}

// GetTotalBalance retrieves the total USD value of all assets in the wallet.
// Returns the total balance as a float64 value.
func (r *Rabby) GetTotalBalance(address string) (float64, error) {
	var resp rabbyModels.TotalBalanceResponse
	if err := r.baseChecker.MakeSimpleRequest("total_balance", "GET", nil, &resp, address); err != nil {
		return 0, err
	}
	return resp.TotalUsdValue, nil
}

// GetUsedChains returns a list of blockchain networks that the wallet has interacted with.
// Results are cached to improve performance.
func (r *Rabby) GetUsedChains(address string) ([]string, error) {
	if chains, ok := r.cache.GetChainsCache(address); ok {
		return chains, nil
	}

	var resp rabbyModels.ChainListResponse
	if err := r.baseChecker.MakeSimpleRequest("used_chains", "GET", nil, &resp, address); err != nil {
		return nil, err
	}

	chains := make([]string, 0, len(resp))
	for _, chain := range resp {
		chains = append(chains, chain.ID)
	}
	r.cache.SetChainsCache(address, chains)
	return chains, nil
}

// GetTokenBalanceList retrieves the list of tokens and their balances for a specific chain.
// Returns a slice of TokenInfo containing token details and balances.
func (r *Rabby) GetTokenBalanceList(address, chain string) ([]*commonModels.TokenInfo, error) {
	var resp rabbyModels.TokenBalanceListResponse
	if err := r.baseChecker.MakeSimpleRequest("token_balance_list", "GET", nil, &resp, address, chain); err != nil {
		return nil, err
	}

	result := make([]*commonModels.TokenInfo, 0, len(resp))
	for _, token := range resp {
		result = append(result, &commonModels.TokenInfo{
			Amount: token.Amount,
			Chain:  token.Chain,
			ID:     token.ID,
			Price:  token.Price,
			Symbol: token.Symbol,
			Name:   token.Name,
		})
	}
	return result, nil
}

// GetProjectAssets retrieves the list of projects and their associated assets for a wallet.
// Returns a slice of ProjectAssets containing project details and their token holdings.
func (r *Rabby) GetProjectAssets(address string) ([]*commonModels.ProjectAssets, error) {
	var resp rabbyModels.ProjectListResponse
	if err := r.baseChecker.MakeSimpleRequest("project_list", "GET", nil, &resp, address); err != nil {
		return nil, err
	}

	return r.extractProjectAssets(resp)
}

// extractProjectAssets processes the raw project list response and converts it to a structured format.
// It filters out projects with no assets and organizes the data into ProjectAssets structure.
func (r *Rabby) extractProjectAssets(resp rabbyModels.ProjectListResponse) ([]*commonModels.ProjectAssets, error) {
	if len(resp) == 0 {
		return make([]*commonModels.ProjectAssets, 0), nil
	}

	projectAssets := make([]*commonModels.ProjectAssets, 0, len(resp))
	for _, project := range resp {
		assets := &commonModels.ProjectAssets{
			ProjectName: project.Name,
			SiteUrl:     project.SiteURL,
			Chain:       project.Chain,
			Assets:      make([]*commonModels.TokenInfo, 0),
		}

		for _, item := range project.PortfolioItemList {
			for _, token := range item.AssetTokenList {
				assets.Assets = append(assets.Assets, &commonModels.TokenInfo{
					Amount: token.Amount,
					Chain:  token.Chain,
					ID:     token.ID,
					Price:  token.Price,
					Symbol: token.Symbol,
				})
			}
		}

		if len(assets.Assets) > 0 {
			projectAssets = append(projectAssets, assets)
		}
	}

	return projectAssets, nil
}
