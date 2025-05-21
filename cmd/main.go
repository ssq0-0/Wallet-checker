package main

import (
	"chief-checker/internal/config/appConfig"
	"chief-checker/internal/usecase"
	"chief-checker/pkg/logger"
	"log"
)

func main() {
	cfg, err := appConfig.LoadConfig("internal/config/appConfig/config.json")
	if err != nil {
		log.Fatalf("failed to load config: %+v", err)
		return
	}

	logLevel := logger.ParseLevel(cfg.LoggerLevel)
	logger.Init(logLevel)

	usecase := usecase.NewUseCase(cfg)
	if err := usecase.Run(); err != nil {
		logger.GlobalLogger.Fatal(err)
		return
	}
}
