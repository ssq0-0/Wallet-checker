package debankModels

type ProjectAssets struct {
	ProjectName string
	SiteUrl     string
	Chain       string
	Assets      []*TokenInfo
}
