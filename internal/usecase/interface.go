// Package usecase provides the application's business logic layer.
// It defines interfaces and implementations for various use cases and handlers.
package usecase

// UseCaseInterface defines the contract for all use cases in the application.
// Each use case represents a specific business operation or workflow.
type UseCaseInterface interface {
	// Run executes the use case's business logic.
	// Returns an error if the operation fails.
	Run() error
}

// HandlerInterface defines the contract for all handlers in the application.
// Handlers are responsible for processing specific types of requests or operations.
type HandlerInterface interface {
	// Handle processes the handler's specific operation.
	// Returns an error if the operation fails.
	Handle() error
}
