// Package appConfig provides application configuration management.
// It handles loading, parsing and validating application configuration from JSON files.
package appConfig

import (
	"chief-checker/pkg/errors"
	"chief-checker/pkg/logger"
	"encoding/json"
	"os"
)

// LoadConfig loads and parses the application configuration from the specified JSON file.
// It uses Viper for configuration management and supports environment variable overrides.
//
// Parameters:
// - path: path to the configuration file
//
// Returns:
// - *Config: parsed configuration structure
// - error: if loading or parsing fails
//
// The configuration file should be in JSON format and contain all required settings.
// Environment variables can override configuration values using the format:
// APP_SETTING_NAME=value
func LoadConfig(path string) (*Config, error) {
	logger.GlobalLogger.Debugf("Loading config from: %s", path)

	data, err := os.ReadFile(path)
	if err != nil {
		logger.GlobalLogger.Errorf("Failed to read config file: %v", err)
		return nil, errors.Wrap(errors.ErrConfigRead, err.Error())
	}

	appConfig := &Config{}
	if err := json.Unmarshal(data, appConfig); err != nil {
		logger.GlobalLogger.Errorf("Failed to unmarshal config: %v", err)
		return nil, errors.Wrap(errors.ErrConfigParse, err.Error())
	}

	if appConfig.Checkers.Debank == nil {
		logger.GlobalLogger.Error("Debank settings are nil")
		return nil, errors.Wrap(errors.ErrValueEmpty, "config is empty")
	}

	return appConfig, nil
}
