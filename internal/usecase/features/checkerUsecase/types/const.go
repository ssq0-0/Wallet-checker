// Package types defines the data structures used throughout the checker system.
// It includes both raw and aggregated data structures for blockchain account information.
package types

// System-wide constants for controlling batch processing and retry behavior
const (
	// BatchSize is the number of results to accumulate before writing
	BatchSize = 5

	// ChainSemaphore is the maximum number of concurrent chain operations
	ChainSemaphore = 20

	// MaxRetries is the maximum number of retry attempts for failed operations
	MaxRetries = 4

	// RetryDelayBase is the base delay in milliseconds between retry attempts
	RetryDelayBase = 500

	// RetryDelayRandom is the maximum random additional delay in milliseconds
	RetryDelayRandom = 100
)
