// Package appConfig provides application configuration management.
// It handles loading, parsing and validating application configuration from JSON files.
package appConfig

import (
	"chief-checker/pkg/errors"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
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
	v := viper.New()

	dir := filepath.Dir(path)
	file := filepath.Base(path)

	configName := strings.TrimSuffix(file, filepath.Ext(file))

	v.SetConfigName(configName)
	v.SetConfigType("json")
	v.AddConfigPath(dir)

	appConfig := &Config{}
	if err := v.ReadInConfig(); err != nil {
		return nil, errors.Wrap(errors.ErrConfigRead, err.Error())
	}

	if err := v.Unmarshal(appConfig); err != nil {
		return nil, errors.Wrap(errors.ErrConfigParse, err.Error())
	}

	if appConfig.Checkers.Debank == nil {
		return nil, errors.Wrap(errors.ErrValueEmpty, "config is empty")
	}

	return appConfig, nil
}
