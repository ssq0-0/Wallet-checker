package httpClient

import (
	"bytes"
	"chief-checker/internal/infrastructure/proxyPool"
	"chief-checker/pkg/logger"
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	utls "github.com/refraction-networking/utls"
	"golang.org/x/net/http2"
)

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
	}
}

// HttpClient implements the HttpClientInterface with support for proxy rotation
// and automatic retries. It handles JSON request/response marshaling and provides HTTP/2 support.
type HttpClient struct {
	client    *http.Client        // HTTP клиент для выполнения запросов
	transport http.RoundTripper   // Транспорт для настройки соединений
	proxyPool proxyPool.ProxyPool // Пул прокси-серверов
	config    Config              // Конфигурация клиента
	mutex     sync.Mutex          // Мьютекс для синхронизации доступа к transport
}

// type contextKey string

// const cancelFuncKey contextKey = "cancelFunc"

// createUTLSDialer создает функцию для установки TLS соединения с использованием uTLS
func createUTLSDialer(serverName string, clientID string) func(ctx context.Context, network, addr string) (net.Conn, error) {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, _, _ := net.SplitHostPort(addr)
		if host == "" {
			host = serverName
		}

		tcpConn, err := (&net.Dialer{Timeout: 10 * time.Second}).DialContext(ctx, network, addr)
		if err != nil {
			return nil, err
		}

		uTlsConfig := &utls.Config{
			ServerName:         host,
			InsecureSkipVerify: false,
		}

		var uTlsConn *utls.UConn
		switch clientID {
		case "Chrome_112":
			uTlsConn = utls.UClient(tcpConn, uTlsConfig, utls.HelloChrome_Auto)
		default:
			uTlsConn = utls.UClient(tcpConn, uTlsConfig, utls.HelloChrome_Auto)
		}

		if err := uTlsConn.Handshake(); err != nil {
			return nil, err
		}

		return uTlsConn, nil
	}
}

// createHTTP2Transport создает транспорт с поддержкой HTTP/2 для uTLS
func createHTTP2Transport(dialTLS func(ctx context.Context, network, addr string) (net.Conn, error)) *http2.Transport {
	return &http2.Transport{
		DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
			ctx := context.Background()
			return dialTLS(ctx, network, addr)
		},
		AllowHTTP:                  false,
		StrictMaxConcurrentStreams: true,
	}
}

// NewHttpClient creates a new HTTP client instance with the specified proxy pool.
// It configures the transport layer with proper TLS and proxy settings.
func NewHttpClient(proxyPool proxyPool.ProxyPool, config Config) (HttpClientInterface, error) {
	// Базовый хост для SNI
	const defaultHost = "api.debank.com" // TODO: вынести в конфиг

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			ServerName:         defaultHost,
			InsecureSkipVerify: false,
		},
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		MaxConnsPerHost:     100,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		DisableKeepAlives:   true,  // Отключаем Keep-Alive для уникальных TLS handshakes
		ForceAttemptHTTP2:   false, // Отключаем HTTP/2 для надежной работы uTLS
	}

	if config.UseUTLS {
		// Настраиваем uTLS
		transport.DialTLSContext = createUTLSDialer(defaultHost, config.UTLSClientID)
	} else if config.RandomizeTLS {
		// Настраиваем случайный TLS для имитации реального браузера
		transport.TLSClientConfig.MinVersion = tls.VersionTLS12
		transport.TLSClientConfig.MaxVersion = tls.VersionTLS13

		// Случайные предпочтения шифров
		cipherSuites := []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
		}
		// Перемешиваем шифры для разных клиентов
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		for i := range cipherSuites {
			j := r.Intn(len(cipherSuites))
			cipherSuites[i], cipherSuites[j] = cipherSuites[j], cipherSuites[i]
		}
		transport.TLSClientConfig.CipherSuites = cipherSuites[:3+r.Intn(3)]
	}

	client := &HttpClient{
		client:    &http.Client{Transport: transport, Timeout: config.Timeout},
		transport: transport,
		proxyPool: proxyPool,
		config:    config,
	}

	return client, nil
}

// RequestWithRetry sends an HTTP request with a JSON payload and automatically retries
// on failures such as rate limiting or proxy issues. It unmarshals the JSON response
// into the provided response body structure.
//
// Parameters:
// - ctx: Context for cancellation and timeouts
// - url: The target URL for the request
// - method: The HTTP method to use (GET, POST, etc.)
// - reqBody: The request body to send (will be JSON encoded)
// - respBody: Pointer to struct where response will be unmarshaled
// - headers: Map of HTTP headers to include in the request
//
// Returns an error if the request fails after all retry attempts.
func (h *HttpClient) RequestWithRetry(ctx context.Context, urlStr, method string, reqBody, respBody interface{}, headers map[string]string) error {
	req, err := h.createRequestWithContext(ctx, urlStr, method, reqBody)
	if err != nil {
		return fmt.Errorf("ошибка создания запроса: %v", err)
	}

	// Устанавливаем ServerName для uTLS, если используется
	if h.config.UseUTLS {
		if parsedURL, err := url.Parse(urlStr); err == nil {
			if transport, ok := h.transport.(*http.Transport); ok {
				transport.DialTLSContext = createUTLSDialer(parsedURL.Hostname(), h.config.UTLSClientID)
			}
		}
	}

	return h.executeWithRetries(req, respBody, headers)
}

// SimpleRequest sends an HTTP request with a JSON payload, similar to RequestWithRetry,
// but without any retry mechanism. It will fail immediately if the request encounters any error.
// Use this when you need to make a single attempt only, or when handling retries at a higher level.
//
// Parameters:
// - ctx: Context for cancellation and timeouts
// - url: The target URL for the request
// - method: The HTTP method to use (GET, POST, etc.)
// - reqBody: The request body to send (will be JSON encoded)
// - respBody: Pointer to struct where response will be unmarshaled
// - headers: Map of HTTP headers to include in the request
//
// Returns an error if the request fails.
func (h *HttpClient) SimpleRequest(ctx context.Context, urlStr, method string, reqBody, respBody interface{}, headers map[string]string) error {
	req, err := h.createRequestWithContext(ctx, urlStr, method, reqBody)
	if err != nil {
		return fmt.Errorf("ошибка создания запроса: %v", err)
	}

	// Устанавливаем ServerName для uTLS, если используется
	if h.config.UseUTLS {
		if parsedURL, err := url.Parse(urlStr); err == nil {
			if transport, ok := h.transport.(*http.Transport); ok {
				transport.DialTLSContext = createUTLSDialer(parsedURL.Hostname(), h.config.UTLSClientID)
			}
		}
	}

	return h.executeWithoutRetries(req, respBody, headers)
}

// createRequestWithContext creates a new HTTP request with the provided context.
// It handles JSON marshaling of the request body if provided.
func (h *HttpClient) createRequestWithContext(ctx context.Context, url, method string, reqBody interface{}) (*http.Request, error) {
	var body io.Reader

	if reqBody != nil {
		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return nil, fmt.Errorf("ошибка маршалинга тела запроса: %v", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания HTTP запроса: %v", err)
	}

	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

// executeWithRetries executes the HTTP request with automatic retries and proxy rotation.
// It handles proxy authentication errors and rate limiting responses.
func (h *HttpClient) executeWithRetries(req *http.Request, respBody interface{}, headers map[string]string) error {
	// Добавляем пользовательские заголовки
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Получаем контекст из запроса
	ctx := req.Context()
	var currentProxy string

	for attempts := 0; attempts < h.config.MaxRetries; attempts++ {
		// Проверяем, не истек ли контекст
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// Обновляем прокси и TLS конфигурацию для каждого запроса
		if h.proxyPool != nil && h.config.UseProxyPool {
			currentProxy = h.proxyPool.GetFreeProxy()
			if currentProxy == "" {
				return fmt.Errorf("нет доступных прокси")
			}

			proxyURL, err := url.Parse(currentProxy)
			if err != nil {
				return fmt.Errorf("неверный формат прокси %s: %v", currentProxy, err)
			}

			h.mutex.Lock()
			if transport, ok := h.transport.(*http.Transport); ok {
				transport.Proxy = http.ProxyURL(proxyURL)
				if h.config.UseUTLS {
					// Обновляем SNI-дилер с текущим хостом
					transport.DialTLSContext = createUTLSDialer(req.URL.Hostname(), h.config.UTLSClientID)
				}
			}
			h.mutex.Unlock()
		}

		// // Добавляем браузерные заголовки, если требуется
		// if h.config.BrowserHeaders {
		// 	h.addBrowserHeaders(req)
		// }

		resp, err := h.client.Do(req)
		if shouldRetry, retryErr := h.handleRequestError(err, currentProxy); shouldRetry {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(h.config.RetryDelay):
				continue
			}
		} else if retryErr != nil {
			return retryErr
		}

		if resp == nil {
			return fmt.Errorf("получен пустой ответ")
		}
		defer resp.Body.Close()

		if shouldRetry, retryErr := h.handleResponseStatus(resp, currentProxy, attempts); shouldRetry {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(h.config.RetryDelay):
				continue
			}
		} else if retryErr != nil {
			return retryErr
		}

		if err := h.parseResponse(resp, respBody); err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("запрос не удался после %d попыток", h.config.MaxRetries)
}

// // addBrowserHeaders добавляет реалистичные заголовки браузера
// func (h *HttpClient) addBrowserHeaders(req *http.Request) {
// 	if h.config.SkipHeaders["accept"] {
// 		return
// 	}

// 	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
// 	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
// 	req.Header.Set("Cache-Control", "no-cache")
// 	req.Header.Set("Pragma", "no-cache")
// 	req.Header.Set("Sec-Fetch-Dest", "document")
// 	req.Header.Set("Sec-Fetch-Mode", "navigate")
// 	req.Header.Set("Sec-Fetch-Site", "none")
// 	req.Header.Set("Sec-Fetch-User", "?1")
// 	req.Header.Set("Upgrade-Insecure-Requests", "1")

// 	if h.config.ClientHints && !h.config.SkipHeaders["sec-ch-ua"] {
// 		req.Header.Set("Sec-CH-UA", `"Not A(Brand";v="99", "Google Chrome";v="121", "Chromium";v="121"`)
// 		req.Header.Set("Sec-CH-UA-Mobile", "?0")
// 		req.Header.Set("Sec-CH-UA-Platform", `"macOS"`)
// 	}
// }

func (h *HttpClient) executeWithoutRetries(req *http.Request, respBody interface{}, headers map[string]string) error {
	// Сначала добавляем пользовательские заголовки
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	ctx := req.Context()
	var currentProxy string

	// Настраиваем прокси и TLS, если используется
	if h.proxyPool != nil && h.config.UseProxyPool {
		// Проверяем, не истек ли контекст
		if ctx.Err() != nil {
			return ctx.Err()
		}

		currentProxy = h.proxyPool.GetFreeProxy()
		if currentProxy != "" {
			proxyURL, err := url.Parse(currentProxy)
			if err != nil {
				return fmt.Errorf("неверный формат прокси %s: %v", currentProxy, err)
			}

			h.mutex.Lock()
			if transport, ok := h.transport.(*http.Transport); ok {
				transport.Proxy = http.ProxyURL(proxyURL)
				if h.config.UseUTLS {
					// Обновляем SNI-дилер с текущим хостом
					transport.DialTLSContext = createUTLSDialer(req.URL.Hostname(), h.config.UTLSClientID)
				}
			}
			h.mutex.Unlock()
		}
	}

	// // Добавляем браузерные заголовки, если требуется
	// if h.config.BrowserHeaders {
	// 	h.addBrowserHeaders(req)
	// }

	// Выполняем запрос
	resp, err := h.client.Do(req)

	// Обрабатываем ошибку запроса
	if err != nil {
		// Если используется прокси и произошла ошибка аутентификации, блокируем прокси
		if h.proxyPool != nil && currentProxy != "" && !h.config.IsRotatingProxy &&
			(strings.Contains(err.Error(), "Proxy Authentication Required") ||
				strings.Contains(err.Error(), "authentication failed") ||
				strings.Contains(err.Error(), "context deadline exceeded")) {

			h.proxyPool.BlockProxy(currentProxy, h.config.BlockTime)
		}
		return fmt.Errorf("ошибка запроса: %v", err)
	}

	// Проверяем наличие ответа
	if resp == nil {
		return fmt.Errorf("получен пустой ответ")
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusProxyAuthRequired {
		if h.proxyPool != nil && currentProxy != "" && !h.config.IsRotatingProxy {
			h.proxyPool.BlockProxy(currentProxy, h.config.BlockTime)
		}
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ошибка статуса: %d (лимит запросов или проблема с прокси): %s", resp.StatusCode, string(body))
	}

	// Разбираем ответ
	return h.parseResponse(resp, respBody)
}

// configureProxy configures the proxy for the current request
func (h *HttpClient) configureProxy(currentProxy *string) error {
	if h.proxyPool == nil || !h.config.UseProxyPool {
		return nil
	}

	*currentProxy = h.proxyPool.GetFreeProxy()
	if *currentProxy == "" {
		return fmt.Errorf("нет доступных прокси")
	}

	proxyURL, err := url.Parse(*currentProxy)
	if err != nil {
		return fmt.Errorf("неверный формат прокси %s: %v", *currentProxy, err)
	}

	// Устанавливаем прокси для http.Transport
	h.mutex.Lock()
	if transport, ok := h.transport.(*http.Transport); ok {
		transport.Proxy = http.ProxyURL(proxyURL)
	}
	h.mutex.Unlock()

	return nil
}

// handleRequestError processes request errors and determines if a retry is needed
func (h *HttpClient) handleRequestError(err error, currentProxy string) (bool, error) {
	if err == nil {
		return false, nil
	}

	if strings.Contains(err.Error(), "Proxy Authentication Required") ||
		strings.Contains(err.Error(), "authentication failed") ||
		strings.Contains(err.Error(), "context deadline exceeded") {

		if h.proxyPool != nil && !h.config.IsRotatingProxy {
			h.proxyPool.BlockProxy(currentProxy, h.config.BlockTime)
		}

		return true, nil
	}

	return false, fmt.Errorf("ошибка запроса: %v", err)
}

// handleResponseStatus processes response statuses and determines if a retry is needed
func (h *HttpClient) handleResponseStatus(resp *http.Response, currentProxy string, attempt int) (bool, error) {
	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusProxyAuthRequired {
		logger.GlobalLogger.Debugf("Лимит запросов/проблемы с прокси. Попытка %d", attempt+1)

		if h.proxyPool != nil && !h.config.IsRotatingProxy {
			h.proxyPool.BlockProxy(currentProxy, h.config.BlockTime)
		}

		return true, nil
	}

	return false, nil
}

// parseResponse handles the HTTP response, including gzip decompression
// and JSON unmarshaling into the provided response body structure.
func (h *HttpClient) parseResponse(resp *http.Response, respBody interface{}) error {
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.GlobalLogger.Debugf("неожиданный статус код: %d, тело: %s", resp.StatusCode, string(body))
		return fmt.Errorf("неожиданный статус код: %d, тело: %s", resp.StatusCode, string(body))
	}

	reader, err := h.getResponseReader(resp)
	if err != nil {
		return err
	}
	defer reader.Close()

	body, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("ошибка чтения тела ответа: %v", err)
	}

	if respBody == nil {
		return nil
	}
	// log.Printf("bodyL: %s", string(body))
	if err := json.Unmarshal(body, respBody); err != nil {
		return fmt.Errorf("ошибка разбора JSON-ответа: %v", err)
	}

	return nil
}

// getResponseReader returns the appropriate reader based on the Content-Encoding header
func (h *HttpClient) getResponseReader(resp *http.Response) (io.ReadCloser, error) {
	if resp.Header.Get("Content-Encoding") != "gzip" {
		return resp.Body, nil
	}

	gzReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать gzip-ридер: %v", err)
	}

	return gzReader, nil
}

// GetTimeout возвращает таймаут HTTP клиента
func (h *HttpClient) GetTimeout() time.Duration {
	return h.client.Timeout
}
