package appConfig

type Config struct {
	Concurrency int      `json:"concurrency" mapstructure:"concurrency"`
	LoggerLevel string   `json:"logger_level" mapstructure:"logger_level"`
	ServerPort  string   `json:"server_port" mapstructure:"server_port"`
	Checkers    Checkers `json:"checkers_params" mapstructure:"checkers_params"`
}

type Checkers struct {
	AddressFilePath string           `json:"address_file_path" mapstructure:"address_file_path"`
	ProxyFilePath   string           `json:"proxy_file_path" mapstructure:"proxy_file_path"`
	Debank          *CheckerSettings `json:"debank" mapstructure:"debank"`
	Rabby           *CheckerSettings `json:"rabby" mapstructure:"rabby"`
}

type CheckerSettings struct {
	BaseURL         string            `json:"base_url" mapstructure:"base_url"`
	Endpoints       map[string]string `json:"endpoints" mapstructure:"endpoints"`
	UseProxyPool    bool              `json:"use_proxy_pool" mapstructure:"use_proxy_pool"`
	RotateProxy     bool              `json:"rotate_proxy" mapstructure:"rotate_proxy"`
	ContextDeadline int               `json:"deadline_request" mapstructure:"deadline_request"`
}
