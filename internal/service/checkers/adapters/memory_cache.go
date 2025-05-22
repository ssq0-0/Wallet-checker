package adapters

import (
	"chief-checker/internal/service/checkers/port"
	"sync"
	"time"
)

type cacheEntry struct {
	data      []string
	timestamp time.Time
}

type userCacheEntry struct {
	headers   map[string]string
	timestamp time.Time
}

// MemoryCache реализует интерфейс Cache с поддержкой TTL и потокобезопасностью.
type MemoryCache struct {
	mu               *sync.RWMutex
	ttl              time.Duration
	ttlUserId        time.Duration
	chainsUsedCache  map[string]cacheEntry
	userHeadersCache map[string]userCacheEntry
}

// NewMemoryCache создает новый экземпляр MemoryCache.
func NewMemoryCache() port.Cache {
	return &MemoryCache{
		mu:               &sync.RWMutex{},
		ttl:              5 * time.Minute,
		ttlUserId:        1 * time.Minute,
		chainsUsedCache:  make(map[string]cacheEntry),
		userHeadersCache: make(map[string]userCacheEntry),
	}
}

// GetChainsCache возвращает кэшированные цепочки для адреса, если они не устарели.
func (c *MemoryCache) GetChainsCache(address string) ([]string, bool) {
	c.mu.RLock()
	entry, exists := c.chainsUsedCache[address]
	c.mu.RUnlock()
	if !exists {
		return nil, false
	}
	if time.Since(entry.timestamp) > c.ttl {
		c.mu.Lock()
		delete(c.chainsUsedCache, address)
		c.mu.Unlock()
		return nil, false
	}
	return entry.data, true
}

// SetChainsCache сохраняет цепочки для адреса в кэш.
func (c *MemoryCache) SetChainsCache(address string, chains []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.chainsUsedCache[address] = cacheEntry{
		data:      chains,
		timestamp: time.Now(),
	}
}

// GetUserHeadersCache возвращает кэшированные заголовки пользователя, если они не устарели.
func (c *MemoryCache) GetUserHeadersCache(address string) (map[string]string, bool) {
	c.mu.RLock()
	entry, exists := c.userHeadersCache[address]
	c.mu.RUnlock()
	if !exists {
		return nil, false
	}
	if time.Since(entry.timestamp) > c.ttl {
		c.mu.Lock()
		delete(c.userHeadersCache, address)
		c.mu.Unlock()
		return nil, false
	}
	hdtCopy := make(map[string]string)
	for k, v := range entry.headers {
		hdtCopy[k] = v
	}
	return hdtCopy, true
}

// SetUserHeadersCache сохраняет заголовки пользователя в кэш.
func (c *MemoryCache) SetUserHeadersCache(address string, headers map[string]string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.userHeadersCache[address] = userCacheEntry{
		headers:   headers,
		timestamp: time.Now(),
	}
}
