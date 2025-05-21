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

type MemoryCache struct {
	mu               *sync.Mutex
	ttl              time.Duration
	ttlUserId        time.Duration
	chainsUsedCache  map[string]cacheEntry
	userHeadersCache map[string]userCacheEntry
}

func NewMemoryCache() port.Cache {
	return &MemoryCache{
		mu:               &sync.Mutex{},
		ttl:              5 * time.Minute,
		ttlUserId:        1 * time.Minute,
		chainsUsedCache:  make(map[string]cacheEntry),
		userHeadersCache: make(map[string]userCacheEntry),
	}
}

func (c *MemoryCache) GetChainsCache(address string) ([]string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.chainsUsedCache[address]
	if !exists {
		return nil, false
	}

	if time.Since(entry.timestamp) > c.ttl {
		delete(c.chainsUsedCache, address)
		return nil, false
	}

	return entry.data, true
}

func (c *MemoryCache) SetChainsCache(address string, chains []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.chainsUsedCache[address] = cacheEntry{
		data:      chains,
		timestamp: time.Now(),
	}
}

func (c *MemoryCache) GetUserHeadersCache(address string) (map[string]string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.userHeadersCache[address]
	if !exists {
		return nil, false
	}

	if time.Since(entry.timestamp) > c.ttl {
		delete(c.userHeadersCache, address)
		return nil, false
	}

	hdtCopy := make(map[string]string)
	for k, v := range entry.headers {
		hdtCopy[k] = v
	}

	return hdtCopy, true
}

func (c *MemoryCache) SetUserHeadersCache(address string, headers map[string]string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.userHeadersCache[address] = userCacheEntry{
		headers:   headers,
		timestamp: time.Now(),
	}
}
