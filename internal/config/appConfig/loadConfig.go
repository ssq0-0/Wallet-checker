package appConfig

import (
	"chief-checker/pkg/errors"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

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
