package appConfig

import (
	"chief-checker/pkg/errors"
	"fmt"
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

	fmt.Printf("Конфиг после загрузки: %+v\n", appConfig)
	fmt.Printf("Debank: %+v\n", appConfig.Checkers.Debank)

	// Если Debank nil, инициализируем его вручную
	if appConfig.Checkers.Debank == nil {
		appConfig.Checkers.Debank = &DebankSettings{
			BaseURL: v.GetString("checkers_params.debank.base_url"),
			Endpoints: map[string]string{
				"user_info":          v.GetString("checkers_params.debank.endpoints.user_info"),
				"used_chains":        v.GetString("checkers_params.debank.endpoints.used_chains"),
				"token_balance_list": v.GetString("checkers_params.debank.endpoints.token_balance_list"),
				"project_list":       v.GetString("checkers_params.debank.endpoints.project_list"),
			},
			RotateProxy:     v.GetBool("checkers_params.debank.rotate_proxy"),
			ProxyFilePath:   v.GetString("checkers_params.debank.proxy_file_path"),
			ContextDeadline: v.GetInt("checkers_params.debank.deadline_request"),
		}
		fmt.Printf("Debank после инициализации: %+v\n", appConfig.Checkers.Debank)
	}

	return appConfig, nil
}
