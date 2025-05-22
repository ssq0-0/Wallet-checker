package requestModels

// RequestParams содержит параметры подписи для запроса.
type RequestParams struct {
	Nonce     string
	Signature string
	Timestamp string
}

// HeaderInfo содержит информацию для формирования заголовка account.
type HeaderInfo struct {
	RandomAt    string `json:"random_at"`
	RandomID    string `json:"random_id"`
	UserAddress string `json:"user_addr"`
}
