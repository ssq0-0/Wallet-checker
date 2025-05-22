// Package storage provides implementations for data persistence and error handling.
package storage

import (
	"chief-checker/internal/usecase/features/checkerUsecase/interfaces"
	"fmt"
	"sync"
)

// memoryErrorCollector implements thread-safe error collection and storage.
// It groups errors by context for organized reporting.
type memoryErrorCollector struct {
	mu     sync.Mutex          // Protects concurrent access to errors map
	errors map[string][]string // Maps context to list of error messages
}

// NewErrorCollector creates a new instance of the error collector.
//
// Returns:
// - interfaces.ErrorCollector: initialized error collector
func NewErrorCollector() interfaces.ErrorCollector {
	return &memoryErrorCollector{
		errors: make(map[string][]string),
	}
}

// SaveError stores an error message with its associated context.
// This method is thread-safe and can be called from multiple goroutines.
//
// Parameters:
// - context: identifier for grouping related errors
// - errorMsg: the error message to store
func (c *memoryErrorCollector) SaveError(context string, errorMsg string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.errors[context] = append(c.errors[context], errorMsg)
}

// WriteErrors writes all collected errors using the provided writer.
// It formats the errors in a hierarchical structure grouped by context.
//
// Parameters:
// - writer: destination for writing formatted errors
//
// Returns:
// - error: if writing fails
func (c *memoryErrorCollector) WriteErrors(writer interfaces.Writer) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.errors) == 0 {
		return writer.Write([]string{"\n\n=== NO PROCESSING ERRORS FOUND ===\n"})
	}

	var lines []string
	lines = append(lines, "\n\n=== PROCESSING ERRORS ===\n")

	for context, errors := range c.errors {
		lines = append(lines, fmt.Sprintf("\n[%s] errors:", context))
		for _, errMsg := range errors {
			lines = append(lines, "  - "+errMsg)
		}
	}

	fmt.Printf("Writing %d error lines to file\n", len(lines))

	return writer.Write(lines)
}

// HasErrors checks if any errors have been collected.
//
// Returns:
// - bool: true if there are collected errors, false otherwise
func (c *memoryErrorCollector) HasErrors() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return len(c.errors) > 0
}
