// Package interfaces defines the contracts for various components
// of the checker system.
package interfaces

import "chief-checker/internal/domain/account"

// TaskScheduler defines the interface for managing concurrent task execution.
// It handles the scheduling and coordination of account checking tasks.
type TaskScheduler interface {
	// Schedule starts processing the provided accounts.
	// Returns a channel that will receive the results of each task.
	Schedule(accounts []*account.Account) <-chan []string

	// Wait blocks until all scheduled tasks are complete.
	Wait()
}
