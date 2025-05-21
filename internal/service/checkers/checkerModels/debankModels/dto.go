package debankModels

type UsedChainsResponse struct {
	Data struct {
		Chains []string `json:"chains"`
	} `json:"data"`
}

type UserResponse struct {
	Data struct {
		User struct {
			Desc struct {
				UsdValue   float64  `json:"usd_value"`
				UsedChains []string `json:"used_chains"`
			} `json:"desc"`
			// Stats struct {
			// 	UsdValue float64 `json:"usd_value"`
			// 	// TopCoins []struct {
			// 	// 	Amount   float64 `json:"amount"`
			// 	// 	ID       string  `json:"id"`
			// 	// 	LogoURL  string  `json:"logo_url"`
			// 	// 	Percent  float64 `json:"percent"`
			// 	// 	Price    float64 `json:"price"`
			// 	// 	Symbol   string  `json:"symbol"`
			// 	// 	UsdValue float64 `json:"usd_value"`
			// 	// } `json:"top_coins"`
			// 	// TopProtocols []struct {
			// 	// 	ChainID  string  `json:"chain_id"`
			// 	// 	ID       string  `json:"id"`
			// 	// 	LogoURL  string  `json:"logo_url"`
			// 	// 	Name     string  `json:"name"`
			// 	// 	Percent  float64 `json:"percent"`
			// 	// 	UsdValue float64 `json:"usd_value"`
			// 	// } `json:"top_protocols"`
			// 	// TopTokens []struct {
			// 	// 	Amount   float64 `json:"amount"`
			// 	// 	ChainID  string  `json:"chain_id"`
			// 	// 	ID       string  `json:"id"`
			// 	// 	LogoURL  string  `json:"logo_url"`
			// 	// 	Percent  float64 `json:"percent"`
			// 	// 	Price    float64 `json:"price"`
			// 	// 	Symbol   string  `json:"symbol"`
			// 	// 	UsdValue float64 `json:"usd_value"`
			// 	// } `json:"top_tokens"`
			// } `json:"stats"`
		} `json:"user"`
	} `json:"data"`
	ErrorCode int `json:"error_code"`
}

type TokenBalanceListResponse struct {
	// CacheSeconds float64     `json:"_cache_seconds"`
	// Seconds      float64     `json:"_seconds"`
	// UseCache     bool        `json:"_use_cache"`
	Data      []TokenInfo `json:"data"`
	ErrorCode int         `json:"error_code"`
}

type TokenInfo struct {
	Amount float64 `json:"amount"`
	// Balance         *big.Int `json:"balance,omitempty"`
	Chain string `json:"chain"`
	// CreditScore     float64  `json:"credit_score,omitempty"`
	// Decimals int `json:"decimals"`
	// DisplaySymbol *string `json:"display_symbol"`
	ID string `json:"id"`
	// IsCore          bool     `json:"is_core"`
	// IsCustom        bool     `json:"is_custom,omitempty"`
	// IsScam          bool     `json:"is_scam,omitempty"`
	// IsSuspicious    bool     `json:"is_suspicious,omitempty"`
	// IsVerified      bool     `json:"is_verified"`
	// IsWallet        bool     `json:"is_wallet"`
	// LogoURL         string   `json:"logo_url"`
	// Name string `json:"name"`
	// OptimizedSymbol string   `json:"optimized_symbol"`
	Price float64 `json:"price"`
	// Price24hChange  float64  `json:"price_24h_change,omitempty"`
	// ProtocolID      string   `json:"protocol_id"`
	Symbol string `json:"symbol"`
	// TimeAt          *float64 `json:"time_at"`
}

type ProjectListResponse struct {
	// CacheSeconds float64 `json:"_cache_seconds"`
	// Seconds      float64 `json:"_seconds"`
	// UseCache     bool    `json:"_use_cache"`
	Data []struct {
		Chain string `json:"chain"`
		// DaoID                 *string `json:"dao_id"`
		// HasSupportedPortfolio bool    `json:"has_supported_portfolio"`
		// ID                    string  `json:"id"`
		// IsTVL                 bool    `json:"is_tvl"`
		// IsVisibleInDefi       *bool   `json:"is_visible_in_defi"`
		// LogoURL               string  `json:"logo_url"`
		Name string `json:"name"`
		// PlatformTokenID       *string `json:"platform_token_id"`
		PortfolioItemList []struct {
			// AssetDict      map[string]float64 `json:"asset_dict"`
			AssetTokenList []TokenInfo `json:"asset_token_list"`
			// Detail         struct {
			// SupplyTokenList []TokenInfo `json:"supply_token_list"`
			// } `json:"detail"`
			// DetailTypes []string `json:"detail_types"`
			Name string `json:"name"`
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
		SiteURL string `json:"site_url"`
		// TagIDs  []string `json:"tag_ids"`
		// TVL     float64  `json:"tvl"`
	} `json:"data"`
	ErrorCode int `json:"error_code"`
}
