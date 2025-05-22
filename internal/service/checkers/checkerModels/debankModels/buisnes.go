package debankModels

// ProjectAssets содержит информацию об активах пользователя в проекте.
type ProjectAssets struct {
	ProjectName string
	SiteUrl     string
	Chain       string
	Assets      []*TokenInfo
}
