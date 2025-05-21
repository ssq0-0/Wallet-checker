package port

type Cache interface {
	GetChainsCache(address string) ([]string, bool)
	SetChainsCache(address string, chains []string)
	GetUserHeadersCache(address string) (map[string]string, bool)
	SetUserHeadersCache(address string, headers map[string]string)
}
