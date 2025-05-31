package processing

import (
	"chief-checker/internal/usecase/features/checkerUsecase/interfaces"
	"chief-checker/internal/usecase/features/checkerUsecase/types"
	"chief-checker/pkg/errors"
	"sync"
)

type tokenCache struct {
	cache map[string]*types.TokenInfo
	mu    sync.RWMutex
}

func newTokenCache() interfaces.TokenCache {
	return &tokenCache{
		cache: make(map[string]*types.TokenInfo),
	}
}

func (c *tokenCache) Update(token *types.TokenInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if info, exists := c.cache[token.Symbol]; exists {
		info.Amount += token.Amount
		info.UsdValue += token.UsdValue
	} else {
		c.cache[token.Symbol] = token
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

type AccountStatCache struct {
	stats map[string]*types.AggregatedData
	mu    sync.RWMutex
}

func newAccountStatCache() interfaces.AccountStatCache {
	return &AccountStatCache{
		stats: make(map[string]*types.AggregatedData),
	}
}

func (a *AccountStatCache) Update(address string, data *types.AggregatedData) error {
	if address == "" || data == nil {
		return errors.Wrap(errors.ErrValueEmpty, "failed to update account cache")
	}

	a.mu.Lock()
	a.stats[address] = data
	a.mu.Unlock()
	return nil
}

func (a *AccountStatCache) GetAllStats() map[string]*types.AggregatedData {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.stats
}
