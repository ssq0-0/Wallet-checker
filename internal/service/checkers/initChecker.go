package checkers

import (
	"chief-checker/internal/config/appConfig"
	"chief-checker/internal/service/checkers/checkerModels/debankModels"
	"errors"
)

type Checker interface {
	GetTotalBalance(address string) (float64, error)
	GetUsedChains(address string) ([]string, error)
	GetTokenBalanceList(address, chain string) (*debankModels.TokenBalanceListResponse, error)
	GetProjectAssets(address string) ([]*debankModels.ProjectAssets, error)
}

func InitChecker(checkerName string, cfg *appConfig.Checkers) (Checker, error) {
	switch checkerName {
	case "debank":
		debankCfg, err := InitDebankConfig(cfg.Debank)
		if err != nil {
			return nil, err
		}
		return NewDebank(debankCfg)
	default:
		return nil, errors.New("checker not found")
	}
}
