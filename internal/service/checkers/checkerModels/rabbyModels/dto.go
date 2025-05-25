package rabbyModels

// TotalBalanceResponse represents the API response containing the total balance of a user in Rabby.
// It includes the total USD value of all assets in the wallet.
type TotalBalanceResponse struct {
	TotalUsdValue float64 `json:"total_usd_value"`
}

// TokenInfo represents information about a token in the user's Rabby wallet.
// It includes details such as amount, chain, price, and token metadata.
type TokenInfo struct {
	Amount float64 `json:"amount"`         // Token balance
	Chain  string  `json:"chain"`          // Blockchain network
	ID     string  `json:"id"`             // Token identifier
	Price  float64 `json:"price"`          // Token price in USD
	Symbol string  `json:"symbol"`         // Token symbol
	Name   string  `json:"name,omitempty"` // Token name (optional)
}

// TokenBalanceListResponse represents a list of tokens and their balances.
type TokenBalanceListResponse []TokenInfo

// ChainListResponse represents a list of blockchain networks.
// Each entry contains the network identifier.
type ChainListResponse []struct {
	ID string `json:"id"` // Network identifier
}

// ProjectListResponse represents a list of projects and their associated assets.
// It includes project details and portfolio information.
type ProjectListResponse []struct {
	Chain string `json:"chain"` // Blockchain network
	// DaoID                 *string `json:"dao_id"`
	// HasSupportedPortfolio bool    `json:"has_supported_portfolio"`
	// ID                    string  `json:"id"`
	// IsTVL                 bool    `json:"is_tvl"`
	// IsVisibleInDefi       *bool   `json:"is_visible_in_defi"`
	// LogoURL               string  `json:"logo_url"`
	Name string `json:"name"` // Project name
	// PlatformTokenID       *string `json:"platform_token_id"`
	PortfolioItemList []struct {
		// AssetDict      map[string]float64 `json:"asset_dict"`
		AssetTokenList []TokenInfo `json:"asset_token_list"` // List of tokens in the portfolio
		// Detail         struct {
		// SupplyTokenList []TokenInfo `json:"supply_token_list"`
		// } `json:"detail"`
		// DetailTypes []string `json:"detail_types"`
		Name string `json:"name"` // Portfolio item name
		// Pool        struct {
		// 	AdapterID  string  `json:"adapter_id"`
		// 	Chain      string  `json:"chain"`
		// 	Controller string  `json:"controller"`
		// 	ID         string  `json:"id"`
		// 	Index      *string `json:"index"`
		// 	ProjectID  string  `json:"project_id"`
		// 	TimeAt     int64   `json:"time_at"`
		// } `json:"pool"`
		// ProxyDetail struct{} `json:"proxy_detail"`
		// Stats       struct {
		// 	AssetUsdValue float64 `json:"asset_usd_value"`
		// 	DebtUsdValue  float64 `json:"debt_usd_value"`
		// 	NetUsdValue   float64 `json:"net_usd_value"`
		// } `json:"stats"`
		// UpdateAt float64 `json:"update_at"`
	} `json:"portfolio_item_list"`
	SiteURL string `json:"site_url"` // Project website URL
	// TagIDs  []string `json:"tag_ids"`
	// TVL     float64  `json:"tvl"`
}
