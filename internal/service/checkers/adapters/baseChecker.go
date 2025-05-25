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

// ApiChecker implements the ApiClient interface for making API requests.
// It provides functionality for making HTTP requests with proper headers,
// caching, and parameter generation.
type ApiChecker struct {
	baseUrl     string
	endpoints   map[string]string
	httpClient  httpInterfaces.HttpClientInterface
	cache       port.Cache
	generator   port.ParamGenerator
	ctxDeadline int
}

// NewApiChecker creates a new instance of ApiChecker with the provided configuration.
// It initializes the client with base URL, endpoints, HTTP client, cache, and parameter generator.
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

// MakeRequest performs an HTTP request to the API and decodes the response into the result parameter.
// It handles URL formatting, header generation, and request execution.
func (d *ApiChecker) MakeRequest(endpointKey, method, path string, payload map[string]string, result interface{}, urlParams ...string) error {
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

// MakeSimpleRequest performs a simplified HTTP request without additional headers.
// It's used for basic API calls that don't require authentication or special headers.
func (d *ApiChecker) MakeSimpleRequest(endpointKey, method string, payload map[string]string, result interface{}, urlParams ...string) error {
	ctx, cancel := d.createCtx()
	defer cancel()

	url, err := d.getEndpoint(endpointKey)
	if err != nil {
		return err
	}

	params := make([]interface{}, len(urlParams))
	for i, v := range urlParams {
		params[i] = v
	}

	formattedURL := fmt.Sprintf(url, params...)
	if err := d.httpClient.SimpleRequest(ctx, formattedURL, method, nil, result, nil); err != nil {
		logger.GlobalLogger.Debugf("failed to make request: %v, url: %s", err, formattedURL)
		return errors.Wrap(err, "failed to make request")
	}

	return nil
}

// getEndpoint retrieves the full URL for the specified endpoint key.
// It handles both absolute and relative URLs.
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

// getHeaders generates the required headers for the API request.
// It includes authentication headers and other necessary metadata.
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

// getCacheHeaders retrieves or generates headers for the request.
// It caches headers for each address to improve performance.
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

// extractAddress extracts the wallet address from the payload.
// It checks multiple possible keys for the address.
func (d *ApiChecker) extractAddress(payload map[string]string) (string, bool) {
	for _, key := range []string{"id", "user_addr"} {
		if address, ok := payload[key]; ok {
			return address, true
		}
	}
	return "", false
}

// createCtx creates a new context with timeout for the request.
func (d *ApiChecker) createCtx() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(d.ctxDeadline)*time.Second)
	return ctx, cancel
}
