package adapters

import (
	"chief-checker/internal/service/checkers/checkerModels/requestModels"
	"chief-checker/internal/service/checkers/port"
	"chief-checker/pkg/errors"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"sort"
	"strings"
	"sync"
	"time"
)

// IDGenerator генерирует криптографически безопасные идентификаторы.
type IDGenerator struct{}

// NewIDGenerator создает новый экземпляр криптографического генератора ID.
func NewIDGenerator() *IDGenerator {
	return &IDGenerator{}
}

// Generate создает криптографически безопасный случайный идентификатор заданной длины.
func (g *IDGenerator) Generate(length int) string {
	b := make([]byte, length/2+length%2)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	id := hex.EncodeToString(b)
	return id[:length]
}

type RequestParams struct {
	Nonce     string
	Signature string
	Timestamp string
}

var (
	multiplier = big.NewInt(6364136223846793005)
	mask       = new(big.Int).Lsh(big.NewInt(1), 64)
	one        = big.NewInt(1)
)

var bigIntPool = sync.Pool{
	New: func() interface{} {
		return new(big.Int)
	},
}

type ParamGeneratorImpl struct {
	IDGenerator    port.IDGenerate
	nonceGenerator *NonceGenerator
	signer         *Signer
}

func NewParamGeneratorImpl(idGenerator port.IDGenerate) *ParamGeneratorImpl {
	return &ParamGeneratorImpl{
		IDGenerator:    idGenerator,
		nonceGenerator: NewNonceGenerator(),
		signer:         NewSigner(),
	}
}

func (p *ParamGeneratorImpl) Generate(payload map[string]string, method, path string) (*requestModels.RequestParams, error) {
	payloadString, err := json.Marshal(payload)
	if err != nil {
		return nil, errors.Wrap(errors.ErrNoCreatedValue, fmt.Sprintf("failed to marshal payload: %v", err.Error()))
	}

	nonce := p.nonceGenerator.generateNonce()
	timestamp := time.Now().UnixNano()
	tsStr := fmt.Sprintf("%d", timestamp)

	queryString := formatQueryString(string(payloadString))

	signature, err := p.signer.sign(method, path, queryString, nonce, tsStr)
	if err != nil {
		return nil, errors.Wrap(errors.ErrNoCreatedValue, fmt.Sprintf("make signature failed: %v", err))
	}
	return &requestModels.RequestParams{
		Nonce:     "n_" + nonce,
		Signature: signature,
		Timestamp: tsStr,
	}, nil
}

type NonceGenerator struct {
	abc   string
	local *big.Int
}

func NewNonceGenerator() *NonceGenerator {
	return &NonceGenerator{
		abc:   "0123456789ABCDEFGHIJKLMNOPQRSTUVWXTZabcdefghiklmnopqrstuvwxyz",
		local: big.NewInt(0),
	}
}

func (p *NonceGenerator) generateNonce() string {
	result := make([]rune, 40)
	for i := 0; i < 40; i++ {
		r := p.pcg32()
		idx := int((float64(r) / 2147483647.0) * 61.0)
		result[i] = rune(p.abc[idx])
	}
	return string(result)
}

func (p *NonceGenerator) pcg32() int64 {
	tmp := bigIntPool.Get().(*big.Int)
	defer bigIntPool.Put(tmp)

	tmp.Mul(p.local, multiplier)
	tmp.Add(tmp, one)
	tmp.Mod(tmp, mask)
	p.local.Set(tmp)

	shifted := bigIntPool.Get().(*big.Int)
	defer bigIntPool.Put(shifted)

	shifted.Rsh(tmp, 33)
	return shifted.Int64()
}

type Signer struct {
	builderPool sync.Pool
}

func NewSigner() *Signer {
	return &Signer{
		builderPool: sync.Pool{
			New: func() interface{} {
				return new(strings.Builder)
			},
		},
	}
}

func (p *Signer) sign(method, path, query, nonce, ts string) (string, error) {
	data1 := make([]byte, 0, len(method)+len(path)+len(query)+2)
	data1 = append(data1, method...)
	data1 = append(data1, '\n')
	data1 = append(data1, path...)
	data1 = append(data1, '\n')
	data1 = append(data1, query...)

	data2 := make([]byte, 0, 20+len(nonce)+len(ts)+2)
	data2 = append(data2, "debank-api\nn_"...)
	data2 = append(data2, nonce...)
	data2 = append(data2, '\n')
	data2 = append(data2, ts...)

	sum1 := sha256.Sum256(data1)
	hash1 := hex.EncodeToString(sum1[:])

	sum2 := sha256.Sum256(data2)
	hash2 := hex.EncodeToString(sum2[:])

	xor1, xor2 := p.xor(hash2)

	h1sum := sha256.Sum256([]byte(xor1 + hash1))

	combined := make([]byte, 0, len(xor2)+len(h1sum))
	combined = append(combined, xor2...)
	combined = append(combined, h1sum[:]...)
	h2sum := sha256.Sum256(combined)

	return hex.EncodeToString(h2sum[:]), nil
}

func (p *Signer) xor(hash string) (string, string) {
	b1 := p.builderPool.Get().(*strings.Builder)
	b2 := p.builderPool.Get().(*strings.Builder)
	defer func() {
		b1.Reset()
		b2.Reset()
		p.builderPool.Put(b1)
		p.builderPool.Put(b2)
	}()

	b1.Grow(len(hash))
	b2.Grow(len(hash))

	for i := 0; i < len(hash); i++ {
		c := hash[i]
		b1.WriteByte(c ^ 54)
		b2.WriteByte(c ^ 92)
	}
	return b1.String(), b2.String()
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
