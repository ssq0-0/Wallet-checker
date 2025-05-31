package usecaseConfig

import "chief-checker/internal/config/appConfig"

type CheckerHandlerConfig struct {
	ThreadsCount         int
	AddressFilePath      string
	CheckerServiceConfig *appConfig.Checkers
}
