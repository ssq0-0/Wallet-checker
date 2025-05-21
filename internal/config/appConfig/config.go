package appConfig

type Config struct {
	Threads     int      `json:"threads"`
	LoggerLevel string   `json:"logger_level"`
	Checkers    Checkers `json:"checkers_params"`
}

type Checkers struct {
	Debank *DebankSettings `json:"debank"`
}

type DebankSettings struct {
	BaseURL         string            `json:"base_url"`
	Endpoints       map[string]string `json:"endpoints"`
	UseProxyPool    bool              `json:"use_proxy_pool"`
	RotateProxy     bool              `json:"rotate_proxy"`
	ProxyFilePath   string            `json:"proxy_file_path"`
	ContextDeadline int               `json:"deadline_request"`
}
