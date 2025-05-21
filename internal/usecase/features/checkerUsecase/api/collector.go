package api

import (
	"chief-checker/internal/domain/account"
	service "chief-checker/internal/service/checkers"
	"chief-checker/internal/service/checkers/checkerModels/debankModels"
	"chief-checker/internal/usecase/features/checkerUsecase/interfaces"
	"chief-checker/internal/usecase/features/checkerUsecase/types"
	"chief-checker/pkg/errors"
	"chief-checker/pkg/logger"
	"context"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/common"
)

type DataCollector struct {
	errorCollector interfaces.ErrorCollector
	checkerService service.Checker
	minUsdAmount   float64
}

func NewDataCollector(
	errorCollector interfaces.ErrorCollector,
	checkerService service.Checker,
	minUsdAmount float64,
) interfaces.DataCollector {
	return &DataCollector{
		errorCollector: errorCollector,
		checkerService: checkerService,
		minUsdAmount:   minUsdAmount,
	}
}

func (c *DataCollector) CollectData(address string) (*types.RawAccountData, error) {
	return c.collectData(context.Background(), &account.Account{Address: common.HexToAddress(address)})
}

func (c *DataCollector) collectData(ctx context.Context, acc *account.Account) (*types.RawAccountData, error) {
	address := acc.Address.Hex()
	logger.GlobalLogger.Infof("Collecting data for address: %s", address)
	totalBalanceResult, err := c.retryFunc(ctx, "GetTotalBalance", address, func() (interface{}, error) {
		return c.checkerService.GetTotalBalance(address)
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get total balance")
	}
	totalBalance, ok := totalBalanceResult.(float64)
	if !ok {
		return nil, errors.Wrap(errors.ErrUnexpectedType, "unexpected type for total balance")
	}

	if totalBalance <= c.minUsdAmount {
		return nil, errors.Wrap(errors.ErrBalanceEstimation, "total balance is less than min usd amount")
	}

	chainData, projectData, err := c.collectChainsAndProjects(ctx, address)
	if err != nil {
		return nil, err
	}
	return &types.RawAccountData{
		TotalBalance: totalBalance,
		ChainsInfo:   chainData,
		ProjectsInfo: projectData,
	}, nil
}

func (c *DataCollector) collectChainsAndProjects(ctx context.Context, address string) (map[string][]*types.TokenChainInfo, []*debankModels.ProjectAssets, error) {
	var wg sync.WaitGroup
	results := make(chan struct {
		chainsInfo   map[string][]*types.TokenChainInfo
		projectsInfo []*debankModels.ProjectAssets
		err          error
	}, 2)

	wg.Add(1)
	go func() {
		defer wg.Done()
		chainsInfo, err := c.getChainInfo(ctx, address)
		results <- struct {
			chainsInfo   map[string][]*types.TokenChainInfo
			projectsInfo []*debankModels.ProjectAssets
			err          error
		}{chainsInfo: chainsInfo, err: err}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		projectsInfo, err := c.getProjectsInfo(ctx, address)
		results <- struct {
			chainsInfo   map[string][]*types.TokenChainInfo
			projectsInfo []*debankModels.ProjectAssets
			err          error
		}{projectsInfo: projectsInfo, err: err}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	var (
		chainsInfo   = make(map[string][]*types.TokenChainInfo)
		projectsInfo = make([]*debankModels.ProjectAssets, 0)
		firstErr     error
	)

	for result := range results {
		if result.err != nil {
			if firstErr == nil {
				firstErr = result.err
			}
			continue
		}

		if result.chainsInfo != nil {
			chainsInfo = result.chainsInfo
		}

		if result.projectsInfo != nil {
			projectsInfo = result.projectsInfo
		}
	}

	if len(chainsInfo) > 0 || len(projectsInfo) > 0 {
		return chainsInfo, projectsInfo, nil
	}

	return nil, nil, firstErr
}

func (c *DataCollector) getChainInfo(ctx context.Context, address string) (map[string][]*types.TokenChainInfo, error) {
	chainsResult, err := c.retryFunc(ctx, "GetUsedChains", address, func() (interface{}, error) {
		return c.checkerService.GetUsedChains(address)
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get used chains")
	}
	chains := chainsResult.([]string)

	if len(chains) == 0 {
		return nil, errors.Wrap(errors.ErrRequestFailed, "no chains found")
	}

	var (
		result = make(map[string][]*types.TokenChainInfo, len(chains))
		wg     sync.WaitGroup
		mu     sync.Mutex
		sem    = make(chan struct{}, types.ChainSemaphore)
		errCh  = make(chan error, len(chains))
	)

	for _, chain := range chains {
		wg.Add(1)
		go func(chain string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() {
				<-sem
			}()

			tokenChainInfoResult, err := c.retryFunc(ctx, fmt.Sprintf("processSingleChain:%s", chain), address, func() (interface{}, error) {
				return c.processSingleChain(ctx, address, chain)
			})

			if err != nil {
				errCh <- errors.Wrap(errors.ErrRequestFailed, fmt.Sprintf("failed to process chain %s: %w", chain, err))
				return
			}

			mu.Lock()
			result[chain] = tokenChainInfoResult.([]*types.TokenChainInfo)
			mu.Unlock()
		}(chain)
	}

	wg.Wait()
	close(errCh)

	var chainErrors []error
	for err := range errCh {
		if err != nil {
			chainErrors = append(chainErrors, err)
		}
	}

	if len(chainErrors) > 0 {
		errorMsg := fmt.Sprintf("%d chains failed to process: %v", len(chainErrors), chainErrors)
		c.errorCollector.SaveError(address, errorMsg)
	}

	c.errorCollector.SaveError(address, fmt.Sprintf("Successfully processed %d chains", len(result)))
	return result, nil
}

func (c *DataCollector) processSingleChain(ctx context.Context, address string, chain string) ([]*types.TokenChainInfo, error) {
	chainData, err := c.checkerService.GetTokenBalanceList(address, chain)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get token balance list")
	}

	result := make([]*types.TokenChainInfo, 0, len(chainData.Data))
	for _, token := range chainData.Data {
		result = append(result, &types.TokenChainInfo{
			Amount:   token.Amount,
			Chain:    token.Chain,
			Contract: token.ID,
			Price:    token.Price,
			Symbol:   token.Symbol,
			UsdValue: token.Amount * token.Price,
		})
	}

	return result, nil
}

func (c *DataCollector) getProjectsInfo(ctx context.Context, address string) ([]*debankModels.ProjectAssets, error) {
	projectsResult, err := c.retryFunc(ctx, "GetProjectAssets", address, func() (interface{}, error) {
		return c.checkerService.GetProjectAssets(address)
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get projects")
	}

	projects := projectsResult.([]*debankModels.ProjectAssets)
	c.errorCollector.SaveError(address, fmt.Sprintf("found %d projects", len(projects)))
	return projects, nil
}
