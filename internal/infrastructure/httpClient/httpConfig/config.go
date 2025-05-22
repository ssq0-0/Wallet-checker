package httpConfig

import "time"

// Config contains settings for the HttpClient
type Config struct {
	Timeout         time.Duration   // Таймаут для каждого HTTP запроса
	MaxRetries      int             // Максимальное количество повторных попыток
	RetryDelay      time.Duration   // Задержка между повторными попытками
	UseProxyPool    bool            // Использовать ли пул прокси
	IsRotatingProxy bool            // Признак ротационного прокси, для которого не нужна блокировка
	BlockTime       time.Duration   // Время блокировки проблемного прокси
	BrowserHeaders  bool            // Использовать ли реалистичные заголовки браузера
	RandomizeTLS    bool            // Рандомизировать ли TLS-отпечаток
	ClientHints     bool            // Отправлять ли Client Hints заголовки
	SkipHeaders     map[string]bool // Заголовки, которые не нужно добавлять автоматически
	UseUTLS         bool            // Использовать ли uTLS для эмуляции браузера
	UTLSClientID    string          // ID клиента uTLS (например, "Chrome_112")
	Servername      string          // Имя сервера для SNI
}

// DefaultConfig returns the default configuration for HttpClient
func DefaultConfig() Config {
	return Config{
		Timeout:         10 * time.Second,
		MaxRetries:      5,
		RetryDelay:      3 * time.Second,
		UseProxyPool:    true,
		IsRotatingProxy: true,
		BlockTime:       10 * time.Second,
		BrowserHeaders:  true,
		RandomizeTLS:    true,
		ClientHints:     true,
		SkipHeaders:     map[string]bool{"sec-ch-ua": true, "sec-ch-ua-mobile": true, "sec-ch-ua-platform": true},
		UseUTLS:         false,
		UTLSClientID:    "Chrome_112",
		Servername:      "api.debank.com",
	}
}
