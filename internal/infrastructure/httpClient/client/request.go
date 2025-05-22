package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

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

func (h *HttpClient) executeWithRetries(req *http.Request, respBody interface{}, headers map[string]string) error {
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	ctx := req.Context()
	var currentProxy string

	for attempts := 0; attempts < h.config.MaxRetries; attempts++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

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
			}
			h.mutex.Unlock()
		}

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

func (h *HttpClient) executeWithoutRetries(req *http.Request, respBody interface{}, headers map[string]string) error {
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	ctx := req.Context()
	var currentProxy string

	if h.proxyPool != nil && h.config.UseProxyPool {
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
			}
			h.mutex.Unlock()
		}
	}

	resp, err := h.client.Do(req)
	if err != nil {
		if h.proxyPool != nil && currentProxy != "" && !h.config.IsRotatingProxy &&
			(strings.Contains(err.Error(), "Proxy Authentication Required") ||
				strings.Contains(err.Error(), "authentication failed") ||
				strings.Contains(err.Error(), "context deadline exceeded")) {

			h.proxyPool.BlockProxy(currentProxy, h.config.BlockTime)
		}
		return fmt.Errorf("ошибка запроса: %v", err)
	}

	if resp == nil {
		return fmt.Errorf("получен пустой ответ")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusProxyAuthRequired {
		if h.proxyPool != nil && currentProxy != "" && !h.config.IsRotatingProxy {
			h.proxyPool.BlockProxy(currentProxy, h.config.BlockTime)
		}
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ошибка статуса: %d (лимит запросов или проблема с прокси): %s", resp.StatusCode, string(body))
	}

	return h.parseResponse(resp, respBody)
}
