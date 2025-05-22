package client

import (
	"chief-checker/pkg/logger"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

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
