package requestModels

type RequestParams struct {
	Nonce     string
	Signature string
	Timestamp string
}

type HeaderInfo struct {
	RandomAt    string `json:"random_at"`
	RandomID    string `json:"random_id"`
	UserAddress string `json:"user_addr"`
}
