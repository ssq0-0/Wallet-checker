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
	MakeRequest(endpointKey, method, path string, payload map[string]string, result interface{}, urlParams ...string) error

	// MakeSimpleRequest performs a simplified HTTP request without additional headers and path.
	// It's used for basic API calls that don't require authentication or special headers.
	//
	// Parameters:
	// - endpointKey: identifier for the endpoint configuration
	// - method: HTTP method (GET, POST, etc.)
	// - payload: request parameters
	// - result: pointer to store the response
	// - urlParams: optional URL parameters for endpoint formatting
	//
	// Returns:
	// - error: if the request fails or response processing fails
	MakeSimpleRequest(endpointKey, method string, payload map[string]string, result interface{}, urlParams ...string) error
}
