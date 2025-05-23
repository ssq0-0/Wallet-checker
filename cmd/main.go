package main

import (
	"chief-checker/internal/config/appConfig"
	"chief-checker/internal/usecase"
	"chief-checker/pkg/logger"
	"path/filepath"
	"runtime"
)

func main() {
	// Инициализируем логгер с уровнем debug для отладки
	logger.Init("debug")

	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	configPath := filepath.Join(basepath, "..", "internal", "config", "appConfig", "config.json")

	// logger.GlobalLogger.Debugf("Base path: %s", basepath)
	// logger.GlobalLogger.Debugf("Config path: %s", configPath)

	cfg, err := appConfig.LoadConfig(configPath)
	if err != nil {
		logger.GlobalLogger.Fatalf("failed to load config: %+v", err)
		return
	}

	// Обновляем уровень логирования из конфигурации
	logLevel := logger.ParseLevel(cfg.LoggerLevel)
	logger.Init(logLevel)

	usecase := usecase.NewUseCase(cfg)
	if err := usecase.Run(); err != nil {
		logger.GlobalLogger.Fatal(err)
		return
	}
}
