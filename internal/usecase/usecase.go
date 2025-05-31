// Package usecase provides the application's business logic layer.
// It defines interfaces and implementations for various use cases and handlers.
package usecase

import (
	"chief-checker/internal/config/appConfig"
	"chief-checker/internal/config/usecaseConfig"
	"chief-checker/internal/service/server"
	"chief-checker/internal/service/server/serverFormater"
	"chief-checker/internal/service/server/serverHandler"
	"chief-checker/internal/service/server/serverInterface"
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
		needServer := u.serverNeed()

		var srv serverInterface.Server
		if needServer {
			handler, err := u.initApiCheckerHandler()
			if err != nil {
				return err
			}

			serverHandler := serverHandler.NewServerHandler(
				handler.GetAggregator(),
				serverFormater.NewFormater(),
			)
			srv = server.NewServerHandler(serverHandler)
			go srv.StartServer(u.config.ServerPort)

			if err := handler.Handle(); err != nil {
				return err
			}

			<-srv.Done()
		} else {
			handler, err := u.initApiCheckerHandler()
			if err != nil {
				return err
			}

			if err := handler.Handle(); err != nil {
				return err
			}
		}

		return nil
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
		ThreadsCount:         u.config.Concurrency,
		CheckerServiceConfig: &u.config.Checkers,
		AddressFilePath:      u.config.Checkers.AddressFilePath,
	}

	return checkerUsecase.CreateSystem(checkerConfig)
}

func (u *UseCase) serverNeed() bool {
	serverNeed, err := selector.SelectServer("Нужен ли сервер? (y/n)")
	if err != nil {
		return false
	}
	return serverNeed == "Да"
}
