// Package adapters provides various adapters for external services and utilities.
package adapters

import (
	"chief-checker/internal/service/checkers/port"
	"sync"
	"time"
)

// cacheEntry represents a single cache entry with data and timestamp.
type cacheEntry struct {
	data      []string
	timestamp time.Time
}

// userCacheEntry represents a cached user headers with timestamp.
type userCacheEntry struct {
	headers   map[string]string
	timestamp time.Time
}

// MemoryCache implements the Cache interface with TTL support and thread safety.
// It provides two types of caching:
// - Chain caching: stores blockchain networks used by addresses
// - User headers caching: stores user-specific headers for API requests
//
// All operations are thread-safe and support automatic cleanup of expired entries.
type MemoryCache struct {
	mu               *sync.RWMutex
	ttl              time.Duration
	ttlUserId        time.Duration
	chainsUsedCache  map[string]cacheEntry
	userHeadersCache map[string]userCacheEntry
}

// NewMemoryCache creates a new instance of MemoryCache.
// It initializes the cache with default TTL values:
// - 5 minutes for chain data
// - 1 minute for user headers
//
// The cache is immediately ready for use after creation.
func NewMemoryCache() port.Cache {
	return &MemoryCache{
		mu:               &sync.RWMutex{},
		ttl:              5 * time.Minute,
		ttlUserId:        1 * time.Minute,
		chainsUsedCache:  make(map[string]cacheEntry),
		userHeadersCache: make(map[string]userCacheEntry),
	}
}

// GetChainsCache retrieves cached chains for the given address.
// Returns:
// - []string: list of cached chains
// - bool: true if cache hit and data is not expired, false otherwise
//
// Thread-safe: can be called concurrently from multiple goroutines.
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

// SetChainsCache stores chains for the given address in cache.
// The data will be available for retrieval until TTL expires.
//
// Thread-safe: can be called concurrently from multiple goroutines.
func (c *MemoryCache) SetChainsCache(address string, chains []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.chainsUsedCache[address] = cacheEntry{
		data:      chains,
		timestamp: time.Now(),
	}
}

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

func (c *MemoryCache) SetUserHeadersCache(address string, headers map[string]string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.userHeadersCache[address] = userCacheEntry{
		headers:   headers,
		timestamp: time.Now(),
	}
}
