// Package port defines the interfaces for external service interactions
// in the checker system. It provides contracts for API clients, caching,
// and parameter generation.
package port

// Cache defines the interface for caching user data and chain information.
// It provides methods for storing and retrieving frequently accessed data
// to improve performance and reduce API calls.
type Cache interface {
	// GetChainsCache retrieves cached chains for an address.
	//
	// Parameters:
	// - address: blockchain address to get chains for
	//
	// Returns:
	// - []string: list of chain identifiers
	// - bool: true if cache hit, false if cache miss
	GetChainsCache(address string) ([]string, bool)

	// SetChainsCache stores chains for an address in the cache.
	//
	// Parameters:
	// - address: blockchain address to store chains for
	// - chains: list of chain identifiers to cache
	SetChainsCache(address string, chains []string)

	// GetUserHeadersCache retrieves cached user headers.
	//
	// Parameters:
	// - address: blockchain address to get headers for
	//
	// Returns:
	// - map[string]string: header key-value pairs
	// - bool: true if cache hit, false if cache miss
	GetUserHeadersCache(address string) (map[string]string, bool)

	// SetUserHeadersCache stores user headers in the cache.
	//
	// Parameters:
	// - address: blockchain address to store headers for
	// - headers: header key-value pairs to cache
	SetUserHeadersCache(address string, headers map[string]string)
}
