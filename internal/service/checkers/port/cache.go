package port

// Cache определяет интерфейс для кэширования данных пользователя и цепочек.
type Cache interface {
	// GetChainsCache возвращает кэшированные цепочки для адреса.
	GetChainsCache(address string) ([]string, bool)
	// SetChainsCache сохраняет цепочки для адреса в кэш.
	SetChainsCache(address string, chains []string)
	// GetUserHeadersCache возвращает кэшированные заголовки пользователя.
	GetUserHeadersCache(address string) (map[string]string, bool)
	// SetUserHeadersCache сохраняет заголовки пользователя в кэш.
	SetUserHeadersCache(address string, headers map[string]string)
}
