package checkers

import (
	"chief-checker/internal/config/appConfig"
	"chief-checker/internal/service/checkers/checkerModels/commonModels"
	"errors"
)

// Checker defines the interface for retrieving user balance and asset information.
// It provides methods to fetch total balance, used chains, token balances, and project assets.
type Checker interface {
	// GetTotalBalance retrieves the total USD value of all assets in the wallet.
	GetTotalBalance(address string) (float64, error)
	// GetUsedChains returns a list of blockchain networks that the wallet has interacted with.
	GetUsedChains(address string) ([]string, error)
	// GetTokenBalanceList retrieves the list of tokens and their balances for a specific chain.
	GetTokenBalanceList(address, chain string) ([]*commonModels.TokenInfo, error)
	// GetProjectAssets retrieves the list of projects and their associated assets for a wallet.
	GetProjectAssets(address string) ([]*commonModels.ProjectAssets, error)
}

// InitChecker initializes a checker instance based on the provided name and configuration.
// It supports multiple checker implementations (debank, rabby) and returns the appropriate instance.
// Returns an error if the checker name is not supported or initialization fails.
func InitChecker(checkerName string, cfg *appConfig.Checkers) (Checker, error) {
	switch checkerName {
	case "debank":
		debankCfg, err := InitDebankConfig(cfg.Debank, cfg.ProxyFilePath)
		if err != nil {
			return nil, err
		}
		return NewDebank(debankCfg)
	case "rabby":
		rabbyCfg, err := InitRabbyConfig(cfg.Rabby, cfg.ProxyFilePath)
		if err != nil {
			return nil, err
		}
		return NewRabby(rabbyCfg)
	default:
		return nil, errors.New("checker not found")
	}
}
