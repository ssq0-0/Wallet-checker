package checkers

import (
	"chief-checker/internal/config/serviceConfig/debankConfig"
	"chief-checker/internal/service/checkers/checkerModels/debankModels"
	"chief-checker/internal/service/checkers/port"
	"chief-checker/pkg/logger"
)

type Debank struct {
	// endpoints   map[string]string
	// httpClient  httpClient.HttpClientInterface
	cache port.Cache
	// generator   *ParamGenerator
	ctxDeadline int
	baseChecker port.ApiClient
}

func NewDebank(cfg *debankConfig.DebankConfig) (*Debank, error) {
	factory := NewFactory(cfg)
	return factory.CreateDebank()
}

func (d *Debank) GetTotalBalance(address string) (float64, error) {
	var resp *debankModels.UserResponse
	if err := d.baseChecker.MakeRequest(
		"user_info",
		"GET",
		"/user",
		map[string]string{"id": address},
		&resp,
		address,
	); err != nil {
		return 0, err
	}

	if len(resp.Data.User.Desc.UsedChains) > 0 {
		d.cache.SetChainsCache(address, resp.Data.User.Desc.UsedChains)
	}

	return resp.Data.User.Desc.UsdValue, nil
}

func (d *Debank) GetUsedChains(address string) ([]string, error) {
	if chains, ok := d.cache.GetChainsCache(address); ok {
		return chains, nil
	}

	var resp *debankModels.UsedChainsResponse
	if err := d.baseChecker.MakeRequest(
		"used_chains",
		"GET",
		"/user/used_chains",
		map[string]string{"id": address},
		&resp,
		address,
	); err != nil {
		return nil, err
	}

	d.cache.SetChainsCache(address, resp.Data.Chains)
	return resp.Data.Chains, nil
}

func (d *Debank) GetTokenBalanceList(address, chain string) (*debankModels.TokenBalanceListResponse, error) {
	var resp *debankModels.TokenBalanceListResponse
	if err := d.baseChecker.MakeRequest(
		"token_balance_list",
		"GET",
		"/token/balance_list",
		map[string]string{
			"user_addr": address,
			"chain":     chain,
		},
		&resp,
		address,
		chain,
	); err != nil {
		return nil, err
	}
	logger.GlobalLogger.Debugf("token balance list: %+v", resp)
	return resp, nil
}

func (d *Debank) GetProjectAssets(address string) ([]*debankModels.ProjectAssets, error) {
	var resp *debankModels.ProjectListResponse
	if err := d.baseChecker.MakeRequest(
		"project_list",
		"GET",
		"/portfolio/project_list",
		map[string]string{"user_addr": address},
		&resp,
		address,
	); err != nil {
		return nil, err
	}

	return d.extractProjectAssets(resp)
}

func (d *Debank) extractProjectAssets(resp *debankModels.ProjectListResponse) ([]*debankModels.ProjectAssets, error) {
	if len(resp.Data) == 0 {
		logger.GlobalLogger.Debugf("project list is empty")
		return make([]*debankModels.ProjectAssets, 0), nil
	}

	projectAssets := make([]*debankModels.ProjectAssets, 0, len(resp.Data))
	for _, project := range resp.Data {
		assets := &debankModels.ProjectAssets{
			ProjectName: project.Name,
			SiteUrl:     project.SiteURL,
			Chain:       project.Chain,
		}

		for _, portfolioItem := range project.PortfolioItemList {
			for _, token := range portfolioItem.AssetTokenList {
				tokenCopy := token
				assets.Assets = append(assets.Assets, &tokenCopy)
			}
		}

		if len(assets.Assets) > 0 {
			projectAssets = append(projectAssets, assets)
		}
	}

	return projectAssets, nil
}
