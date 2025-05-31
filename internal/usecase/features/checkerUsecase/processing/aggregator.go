package processing

import (
	"chief-checker/internal/service/checkers/checkerModels/commonModels"
	"chief-checker/internal/usecase/features/checkerUsecase/interfaces"
	"chief-checker/internal/usecase/features/checkerUsecase/types"
	"chief-checker/pkg/logger"
	"sync"
	"sync/atomic"
)

type DataAggregatorImpl struct {
	tokenCache       interfaces.TokenCache
	accountCacheData interfaces.AccountStatCache
	minUsdAmount     float64
	statsCounter     atomic.Int32
}

func NewDataAggregator(minUsdAmount float64) interfaces.DataAggregator {
	return &DataAggregatorImpl{
		tokenCache:       newTokenCache(),
		accountCacheData: newAccountStatCache(),
		minUsdAmount:     minUsdAmount,
	}
}

func (a *DataAggregatorImpl) SetMinUsdAmount(amount float64) {
	a.minUsdAmount = amount
}

func (a *DataAggregatorImpl) AggregateAccountData(address string, data *types.RawAccountData) (*types.AggregatedData, error) {
	if data == nil || data.TotalBalance < a.minUsdAmount {
		return nil, nil
	}

	a.statsCounter.Add(1)

	var wg sync.WaitGroup
	var chainData map[string][]*types.TokenInfo
	var projectData []*types.ProjectInfo

	wg.Add(2)

	go func() {
		defer wg.Done()
		chainData = a.chainAggregate(data.ChainsInfo)
	}()

	go func() {
		defer wg.Done()
		projectData = a.projectDataAggregate(data.ProjectsInfo)
	}()

	wg.Wait()

	aggrData := &types.AggregatedData{
		Address:      address,
		TotalBalance: data.TotalBalance,
		ChainData:    chainData,
		ProjectData:  projectData,
	}

	if err := a.accountCacheData.Update(address, aggrData); err != nil {
		logger.GlobalLogger.Debugf("failed to save accoun data to cache: %v", err)
	}

	return aggrData, nil
}

func (a *DataAggregatorImpl) projectDataAggregate(data []*commonModels.ProjectAssets) []*types.ProjectInfo {
	projectData := make([]*types.ProjectInfo, 0, len(data))

	for _, project := range data {
		var projectTokens []*types.TokenInfo

		if len(project.Assets) > 0 {
			projectTokens = make([]*types.TokenInfo, 0, len(project.Assets))
		}

		for _, token := range project.Assets {
			usdValue := token.Amount * token.Price
			if usdValue < a.minUsdAmount {
				continue
			}

			tokenInfo := &types.TokenInfo{
				Symbol:   token.Symbol,
				Amount:   token.Amount,
				UsdValue: usdValue,
				Chain:    project.Chain,
				Price:    token.Price,
			}
			projectTokens = append(projectTokens, tokenInfo)

			a.tokenCache.Update(tokenInfo)
		}

		if len(projectTokens) > 0 {
			projectData = append(projectData, &types.ProjectInfo{
				Name:    project.ProjectName,
				SiteUrl: project.SiteUrl,
				Chain:   project.Chain,
				Tokens:  projectTokens,
			})
		}
	}
	return projectData
}

func (a *DataAggregatorImpl) chainAggregate(data map[string][]*types.TokenChainInfo) map[string][]*types.TokenInfo {
	chainData := make(map[string][]*types.TokenInfo, len(data))

	for chain, tokens := range data {
		chainTokens := make([]*types.TokenInfo, 0, len(tokens))

		for _, token := range tokens {
			if token.UsdValue < a.minUsdAmount {
				continue
			}

			tokenInfo := &types.TokenInfo{
				Symbol:   token.Symbol,
				Amount:   token.Amount,
				UsdValue: token.UsdValue,
				Contract: token.Contract,
				Chain:    token.Chain,
				Price:    token.Price,
			}
			chainTokens = append(chainTokens, tokenInfo)

			a.tokenCache.Update(tokenInfo)
		}
		if len(chainTokens) > 0 {
			chainData[chain] = chainTokens
		}
	}
	return chainData
}

func (a *DataAggregatorImpl) GetGlobalStats() *types.GlobalStats {
	tokenStats := a.tokenCache.GetAll()
	var totalUsdValue float64

	for _, token := range tokenStats {
		totalUsdValue += token.UsdValue
	}

	return &types.GlobalStats{
		TotalAccounts: a.statsCounter.Load(),
		TokenStats:    tokenStats,
		TotalUsdValue: totalUsdValue,
	}
}

func (a *DataAggregatorImpl) GetAllStats() map[string]*types.AggregatedData {
	return a.accountCacheData.GetAllStats()

}
