package commonModels

// TokenInfo представляет общую информацию о токене, используемую всеми сервисами
type TokenInfo struct {
	Amount float64 `json:"amount"`
	Chain  string  `json:"chain"`
	ID     string  `json:"id"`
	Price  float64 `json:"price"`
	Symbol string  `json:"symbol"`
	Name   string  `json:"name,omitempty"`
}

// ProjectAsset содержит информацию об активах пользователя в проекте
type ProjectAssets struct {
	ProjectName string
	SiteUrl     string
	Chain       string
	Assets      []*TokenInfo
}
