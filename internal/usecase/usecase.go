// Package usecase provides the application's business logic layer.
// It defines interfaces and implementations for various use cases and handlers.
package usecase

import (
	"chief-checker/internal/config/appConfig"
	"chief-checker/internal/config/usecaseConfig"
	"chief-checker/internal/usecase/features/checkerUsecase"
	"chief-checker/internal/usecase/selector"
)

// UseCase represents the main application use case that coordinates
// different features and services based on user selection.
type UseCase struct {
	config *appConfig.Config // Application configuration
}

// NewUseCase creates a new instance of UseCase with the provided configuration.
//
// Parameters:
// - cfg: application configuration
//
// Returns:
// - *UseCase: initialized use case instance
func NewUseCase(cfg *appConfig.Config) *UseCase {
	return &UseCase{
		config: cfg,
	}
}

// Run executes the main application workflow.
// It presents a service selection menu to the user and handles their choice.
//
// Returns:
// - error: if any operation fails during execution
func (u *UseCase) Run() error {
	userSelect, err := selector.SelectService("Выберите сервис:")
	if err != nil {
		return err
	}

	switch userSelect {
	case "API Checker":
		handler, err := u.initApiCheckerHandler()
		if err != nil {
			return err
		}
		return handler.Handle()
	}

	return nil
}

// initApiCheckerHandler initializes the API checker handler with configuration.
// It sets up the checker service with thread count and service-specific settings.
//
// Returns:
// - *checkerUsecase.CheckerHandler: initialized checker handler
// - error: if initialization fails
func (u *UseCase) initApiCheckerHandler() (*checkerUsecase.CheckerHandler, error) {
	checkerConfig := &usecaseConfig.CheckerHandlerConfig{
		ThreadsCount:         u.config.Threads,
		CheckerServiceConfig: &u.config.Checkers,
	}

	return checkerUsecase.CreateSystem(checkerConfig)
}
