package usecaseConfig

import "chief-checker/internal/config/appConfig"

type CheckerHandlerConfig struct {
	ThreadsCount         int
	CheckerServiceConfig *appConfig.Checkers
}
