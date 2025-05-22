// Package interfaces defines the contracts for various components
// of the checker system.
package interfaces

// ErrorCollector defines the interface for collecting and managing errors.
// It provides functionality to save, check, and write error messages.
type ErrorCollector interface {
	// SaveError stores an error message with its context.
	SaveError(context string, errorMsg string)

	// WriteErrors writes all collected errors using the provided writer.
	// Returns an error if writing fails.
	WriteErrors(writer Writer) error

	// HasErrors checks if any errors have been collected.
	HasErrors() bool
}

// Writer defines the interface for writing output data.
// It provides basic file operations for writing and closing.
type Writer interface {
	// Write writes multiple lines to the output.
	// Returns an error if writing fails.
	Write(lines []string) error

	// Close finalizes the writing and releases resources.
	// Returns an error if closing fails.
	Close() error
}
