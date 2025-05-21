package adapters

import (
	"chief-checker/internal/infrastructure/wasmClient"
	"chief-checker/internal/service/checkers/checkerModels/requestModels"
	"chief-checker/internal/service/checkers/port"
	"chief-checker/pkg/errors"
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"
)

type IDGenerator struct {
	chars string
	rand  *rand.Rand
	mu    sync.Mutex
}

func NewIDGenerator() *IDGenerator {
	return &IDGenerator{
		chars: "abcdef0123456789",
		rand:  rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (g *IDGenerator) Generate(length int) string {
	g.mu.Lock()
	defer g.mu.Unlock()
	result := make([]byte, length)
	for i := range result {
		result[i] = g.chars[g.rand.Intn(len(g.chars))]
	}
	return string(result)
}

type RequestParams struct {
	Nonce     string
	Signature string
	Timestamp string
}

type ParamGenerator struct {
	IDGenerator port.IDGenerate
	Wasm        wasmClient.Wasm
}

func NewParamGenerator(idGenerator port.IDGenerate, wasm wasmClient.Wasm) *ParamGenerator {
	return &ParamGenerator{
		IDGenerator: idGenerator,
		Wasm:        wasm,
	}
}

func (p *ParamGenerator) Generate(payload map[string]string, method, path string) (*requestModels.RequestParams, error) {
	payloadString, err := json.Marshal(payload)
	if err != nil {
		return nil, errors.Wrap(errors.ErrNoCreatedValue, fmt.Sprintf("failed to marshal payload: %v", err.Error()))
	}

	nonce, err := p.Wasm.GenerateNonce()
	if err != nil {
		return nil, errors.Wrap(errors.ErrNoCreatedValue, fmt.Sprintf("generate nonce failed: %v", err.Error()))
	}

	timestamp := time.Now().UnixNano()
	tsStr := fmt.Sprintf("%d", timestamp)

	method = strings.ToUpper(method)
	queryString := formatQueryString(string(payloadString))

	signature, err := p.Wasm.MakeSignature(method, path, queryString, nonce, tsStr)
	if err != nil {
		return nil, errors.Wrap(errors.ErrNoCreatedValue, fmt.Sprintf("make signature failed: %v", err))
	}
	// logger.GlobalLogger.Debugf("[DATA] query: %s, signature: %s, nonce: %s, timestamp: %s", queryString, signature, nonce, tsStr)
	return &requestModels.RequestParams{
		Nonce:     nonce,
		Signature: signature,
		Timestamp: tsStr,
	}, nil
}

func formatQueryString(jsonStr string) string {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return ""
	}

	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	pairs := make([]string, 0, len(keys))
	for _, k := range keys {
		var valueStr string
		if data[k] == nil {
			valueStr = ""
		} else {
			switch v := data[k].(type) {
			case string:
				valueStr = v
			case float64, int, int64, bool:
				valueStr = fmt.Sprintf("%v", v)
			default:
				valueBytes, _ := json.Marshal(v)
				valueStr = string(valueBytes)
			}
		}
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, valueStr))
	}

	return strings.Join(pairs, "&")
}
