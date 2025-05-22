// Package interfaces defines the contracts for various components
// of the checker system.
package interfaces

import service "chief-checker/internal/service/checkers"

// CheckerFactory defines the interface for creating checker services.
// It abstracts the creation of specific blockchain balance checker implementations.
type CheckerFactory interface {
	// CreateChecker creates a checker service by name.
	// Returns the created checker or an error if creation fails.
	CreateChecker(name string) (service.Checker, error)
}
