package processing

import (
	"chief-checker/internal/service/checkers/checkerModels/commonModels"
	"chief-checker/internal/usecase/features/checkerUsecase/interfaces"
	"chief-checker/internal/usecase/features/checkerUsecase/types"
	"sync"
	"sync/atomic"
)

type DataAggregatorImpl struct {
	tokenCache   interfaces.TokenCache
	minUsdAmount float64
	statsCounter atomic.Int32
}

func NewDataAggregator(minUsdAmount float64) interfaces.DataAggregator {
	return &DataAggregatorImpl{
		tokenCache:   newTokenCache(),
		minUsdAmount: minUsdAmount,
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

	return &types.AggregatedData{
		Address:      address,
		TotalBalance: data.TotalBalance,
		ChainData:    chainData,
		ProjectData:  projectData,
	}, nil
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

			a.tokenCache.Update(token.Symbol, token.Amount, usdValue)
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

			a.tokenCache.Update(token.Symbol, token.Amount, token.UsdValue)
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

type tokenCache struct {
	cache map[string]*types.TokenInfo
	mu    sync.RWMutex
}

func newTokenCache() interfaces.TokenCache {
	return &tokenCache{
		cache: make(map[string]*types.TokenInfo),
	}
}

func (c *tokenCache) Update(symbol string, amount, usdValue float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if info, exists := c.cache[symbol]; exists {
		info.Amount += amount
		info.UsdValue += usdValue
	} else {
		c.cache[symbol] = &types.TokenInfo{
			Symbol:   symbol,
			Amount:   amount,
			UsdValue: usdValue,
		}
	}
}

func (c *tokenCache) Get(symbol string) *types.TokenInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cache[symbol]
}

func (c *tokenCache) GetAll() map[string]*types.TokenInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]*types.TokenInfo, len(c.cache))
	for k, v := range c.cache {
		result[k] = &types.TokenInfo{
			Symbol:   v.Symbol,
			Amount:   v.Amount,
			UsdValue: v.UsdValue,
			Contract: v.Contract,
			Chain:    v.Chain,
			Price:    v.Price,
		}
	}
	return result
}
