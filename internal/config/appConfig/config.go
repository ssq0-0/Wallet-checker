package appConfig

type Config struct {
	Concurrency int      `json:"concurrency" mapstructure:"concurrency"`
	LoggerLevel string   `json:"logger_level" mapstructure:"logger_level"`
	Checkers    Checkers `json:"checkers_params" mapstructure:"checkers_params"`
}

type Checkers struct {
	Debank *DebankSettings `json:"debank" mapstructure:"debank"`
}

type DebankSettings struct {
	BaseURL         string            `json:"base_url" mapstructure:"base_url"`
	Endpoints       map[string]string `json:"endpoints" mapstructure:"endpoints"`
	UseProxyPool    bool              `json:"use_proxy_pool" mapstructure:"use_proxy_pool"`
	RotateProxy     bool              `json:"rotate_proxy" mapstructure:"rotate_proxy"`
	ProxyFilePath   string            `json:"proxy_file_path" mapstructure:"proxy_file_path"`
	ContextDeadline int               `json:"deadline_request" mapstructure:"deadline_request"`
}
