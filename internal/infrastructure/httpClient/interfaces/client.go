package interfaces

import (
	"context"
)

// Client определяет интерфейс для HTTP клиента
type Client interface {
	// Request выполняет HTTP запрос с поддержкой повторных попыток
	Request(ctx context.Context, url, method string, reqBody, respBody interface{}, headers map[string]string) error
}
