package port

type ApiClient interface {
	MakeRequest(endpointKey string, method string, path string, payload map[string]string, result interface{}, urlParams ...string) error
}
