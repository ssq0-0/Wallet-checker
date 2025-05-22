package adapters

import (
	"chief-checker/internal/infrastructure/httpClient/httpInterfaces"
	"chief-checker/internal/service/checkers/checkerModels/requestModels"
	"chief-checker/internal/service/checkers/port"
	"chief-checker/pkg/errors"
	"chief-checker/pkg/logger"
	"chief-checker/pkg/useragent"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ApiChecker реализует интерфейс ApiClient для работы с API.
type ApiChecker struct {
	baseUrl     string
	endpoints   map[string]string
	httpClient  httpInterfaces.HttpClientInterface
	cache       port.Cache
	generator   port.ParamGenerator
	ctxDeadline int
}

// NewApiChecker создает новый экземпляр ApiChecker.
func NewApiChecker(baseUrl string, endpoints map[string]string, httpClient httpInterfaces.HttpClientInterface, cache port.Cache, generator port.ParamGenerator, ctxDeadline int) port.ApiClient {
	return &ApiChecker{
		baseUrl:     baseUrl,
		endpoints:   endpoints,
		httpClient:  httpClient,
		cache:       cache,
		generator:   generator,
		ctxDeadline: ctxDeadline,
	}
}

// MakeRequest выполняет HTTP-запрос к API и декодирует результат в result.
func (d *ApiChecker) MakeRequest(endpointKey string, method string, path string, payload map[string]string, result interface{}, urlParams ...string) error {
	ctx, cancel := d.createCtx()
	defer cancel()

	url, err := d.getEndpoint(endpointKey)
	if err != nil {
		return err
	}

	headers, err := d.getHeaders(payload, method, path)
	if err != nil {
		return errors.Wrap(err, "failed to get headers")
	}

	params := make([]interface{}, len(urlParams))
	for i, v := range urlParams {
		params[i] = v
	}

	formattedURL := fmt.Sprintf(url, params...)
	if err := d.httpClient.SimpleRequest(ctx, formattedURL, method, nil, result, headers); err != nil {
		logger.GlobalLogger.Debugf("failed to make request: %v, url: %s", err, formattedURL)
		return errors.Wrap(err, "failed to make request")
	}

	return nil
}

func (d *ApiChecker) getEndpoint(key string) (string, error) {
	url, ok := d.endpoints[key]
	if !ok {
		return "", errors.Wrap(errors.ErrNoCreatedValue, fmt.Sprintf("endpoint %s not found", key))
	}
	if !strings.HasPrefix(url, "http") {
		url = d.baseUrl + url
	}
	return url, nil
}

func (d *ApiChecker) getHeaders(payload map[string]string, method, path string) (map[string]string, error) {
	hashParams, err := d.generator.Generate(payload, method, path)
	if err != nil {
		return nil, errors.Wrap(errors.ErrNoCreatedValue, fmt.Sprintf("failed to generate request params: %s", err.Error()))
	}

	accountHeader, err := d.getCacheHeaders(payload)
	if err != nil {
		return nil, errors.Wrap(errors.ErrNoCreatedValue, fmt.Sprintf("failed to fetch account headers: %s", err.Error()))
	}

	accountHeader["x-api-nonce"] = hashParams.Nonce
	accountHeader["x-api-sign"] = hashParams.Signature
	accountHeader["x-api-ts"] = hashParams.Timestamp

	return accountHeader, nil
}

func (d *ApiChecker) getCacheHeaders(payload map[string]string) (map[string]string, error) {
	address, ok := d.extractAddress(payload)
	if !ok {
		return nil, errors.Wrap(errors.ErrAddressGeneration, "failed to extract address")
	}

	headers, ok := d.cache.GetUserHeadersCache(address)
	if ok {
		return headers, nil
	}

	timestamp := time.Now().UnixNano()
	timestampStr := strconv.FormatInt(timestamp, 10)
	headerInfo := &requestModels.HeaderInfo{
		RandomAt:    timestampStr,
		RandomID:    NewIDGenerator().Generate(32),
		UserAddress: address,
	}
	accountHeaderBytes, err := json.Marshal(headerInfo)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal account header")
	}

	newCacheHeaders := map[string]string{
		"accept":             "*/*",
		"accept-language":    useragent.GetRandomLanguageString(),
		"origin":             d.baseUrl,
		"referer":            d.baseUrl,
		"source":             "web",
		"x-api-ver":          "v2",
		"user-agent":         useragent.GetPlatformSpecificUserAgent(),
		"sec-ch-ua":          useragent.GetSecChUa(),
		"sec-ch-ua-platform": useragent.GetPlatform(),
		"account":            string(accountHeaderBytes),
	}

	d.cache.SetUserHeadersCache(address, newCacheHeaders)
	return newCacheHeaders, nil
}

func (d *ApiChecker) extractAddress(payload map[string]string) (string, bool) {
	for _, key := range []string{"id", "user_addr"} {
		if address, ok := payload[key]; ok {
			return address, true
		}
	}
	return "", false
}

func (d *ApiChecker) createCtx() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(d.ctxDeadline)*time.Second)
	return ctx, cancel
}
