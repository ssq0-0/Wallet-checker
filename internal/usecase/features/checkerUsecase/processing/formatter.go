package processing

import (
	"chief-checker/internal/usecase/features/checkerUsecase/interfaces"
	"chief-checker/internal/usecase/features/checkerUsecase/types"
	"chief-checker/pkg/utils"
	"fmt"
	"sort"
	"strings"
)

type TextFormatter struct {
	minUsdAmount float64
}

func NewTextFormatter(minUsdAmount float64) interfaces.Formatter {
	return &TextFormatter{
		minUsdAmount: minUsdAmount,
	}
}

func (f *TextFormatter) FormatAccountData(data *types.AggregatedData) ([]string, error) {
	if data == nil {
		return nil, nil
	}

	var result []string
	result = append(result, "--------------------------------------------------------------------------------")
	result = append(result, fmt.Sprintf("ADDRESS: %s\nTOTAL BALANCE: %s", data.Address, utils.FormatFloat(data.TotalBalance, 2)))

	result = append(result, "################################################################################")
	result = append(result, "BALANCE BY CHAINS:")

	chainKeys := make([]string, 0, len(data.ChainData))
	for chain := range data.ChainData {
		chainKeys = append(chainKeys, chain)
	}
	sort.Strings(chainKeys)

	for _, chain := range chainKeys {
		tokens := data.ChainData[chain]
		chainInfo := f.formatTokensWithPrefix(chain+":", tokens, func(token *types.TokenInfo) string {
			contractInfo := ""
			if token.Contract != "" && token.Contract != chain {
				contractInfo = fmt.Sprintf(" [%s]", token.Contract)
			}
			return fmt.Sprintf("%s: %s ($%s)%s",
				token.Symbol,
				utils.FormatFloat(token.Amount, 8),
				utils.FormatFloat(token.UsdValue, 2),
				contractInfo)
		})

		if chainInfo != "" {
			result = append(result, chainInfo)
		}
	}

	result = append(result, "################################################################################")
	result = append(result, "BALANCE BY PROJECTS:")

	sort.Slice(data.ProjectData, func(i, j int) bool {
		return data.ProjectData[i].Name < data.ProjectData[j].Name
	})

	for _, project := range data.ProjectData {
		projectInfo := f.formatTokensWithPrefix(project.Name+":", project.Tokens, func(token *types.TokenInfo) string {
			return fmt.Sprintf("%s: %s ($%s) [%s][%s]",
				token.Symbol,
				utils.FormatFloat(token.Amount, 8),
				utils.FormatFloat(token.UsdValue, 2),
				project.SiteUrl,
				project.Chain)
		})

		if projectInfo != "" {
			result = append(result, projectInfo)
		}
	}

	result = append(result, "--------------------------------------------------------------------------------")
	return result, nil
}

func (f *TextFormatter) FormatGlobalStats(stats *types.GlobalStats) ([]string, error) {
	if stats == nil {
		return nil, nil
	}

	var result []string
	result = append(result, "--------------------------------------------------------------------------------")
	result = append(result, "Global Statistics:")

	tokens := f.prepareTokensForDisplay(stats.TokenStats)

	for _, token := range tokens {
		line := fmt.Sprintf("  %s: %s ($%s)",
			token.Symbol,
			utils.FormatFloat(token.Amount, 8),
			utils.FormatFloat(token.UsdValue, 2))
		result = append(result, line)
	}

	result = append(result, fmt.Sprintf("\nTotal accounts processed: %d", stats.TotalAccounts))
	result = append(result, fmt.Sprintf("Total USD value: $%s", utils.FormatFloat(stats.TotalUsdValue, 2)))
	result = append(result, "--------------------------------------------------------------------------------")

	return result, nil
}

func (f *TextFormatter) formatTokensWithPrefix(prefix string, tokens []*types.TokenInfo, tokenFormatter func(*types.TokenInfo) string) string {
	if len(tokens) == 0 {
		return ""
	}

	sort.Slice(tokens, func(i, j int) bool {
		return tokens[i].UsdValue > tokens[j].UsdValue
	})

	var sb strings.Builder
	sb.WriteString(prefix)

	for _, token := range tokens {
		if token.UsdValue < f.minUsdAmount {
			continue
		}

		sb.WriteString("\n  ")
		sb.WriteString(tokenFormatter(token))
	}

	if sb.Len() <= len(prefix) {
		return ""
	}

	return sb.String()
}

func (f *TextFormatter) prepareTokensForDisplay(tokenStats map[string]*types.TokenInfo) []*types.TokenInfo {
	tokens := make([]*types.TokenInfo, 0, len(tokenStats))

	for _, info := range tokenStats {
		if info.UsdValue >= f.minUsdAmount {
			tokens = append(tokens, info)
		}
	}

	sort.Slice(tokens, func(i, j int) bool {
		return tokens[i].UsdValue > tokens[j].UsdValue
	})

	return tokens
}
