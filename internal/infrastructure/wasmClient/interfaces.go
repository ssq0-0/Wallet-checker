package wasmClient

type Wasm interface {
	GenerateNonce() (string, error)
	MakeSignature(method, urlPath, queryString, nonce, tsStr string) (string, error)
	Close() error
}
