package usecase

import (
	"chief-checker/internal/config/appConfig"
	"chief-checker/internal/config/usecaseConfig"
	"chief-checker/internal/usecase/features/checkerUsecase"
	"chief-checker/internal/usecase/selector"
)

type UseCase struct {
	config *appConfig.Config
}

func NewUseCase(cfg *appConfig.Config) *UseCase {
	return &UseCase{
		config: cfg,
	}
}

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

func (u *UseCase) initApiCheckerHandler() (*checkerUsecase.CheckerHandler, error) {
	checkerConfig := &usecaseConfig.CheckerHandlerConfig{
		ThreadsCount:         u.config.Threads,
		CheckerServiceConfig: &u.config.Checkers,
	}

	return checkerUsecase.CreateSystem(checkerConfig)
}
