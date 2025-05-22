// Package port defines the interfaces for external service interactions
// in the checker system. It provides contracts for API clients, caching,
// and parameter generation.
package port

// ApiClient defines the interface for making HTTP requests to external services.
// It abstracts the details of HTTP communication and request signing.
type ApiClient interface {
	// MakeRequest performs an HTTP request to the specified endpoint.
	//
	// Parameters:
	// - endpointKey: identifier for the endpoint configuration
	// - method: HTTP method (GET, POST, etc.)
	// - path: endpoint path
	// - payload: request parameters
	// - result: pointer to store the response
	// - urlParams: optional URL parameters
	//
	// Returns:
	// - error: if the request fails or response processing fails
	MakeRequest(endpointKey string, method string, path string, payload map[string]string, result interface{}, urlParams ...string) error
}
