package httpInterfaces

import (
	"context"
)

type HttpClientInterface interface {
	RequestWithRetry(ctx context.Context, url, method string, reqBody, respBody interface{}, headers map[string]string) error
	SimpleRequest(ctx context.Context, url, method string, reqBody, respBody interface{}, headers map[string]string) error
}
