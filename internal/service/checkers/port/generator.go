// Package port defines the interfaces for external service interactions
// in the checker system. It provides contracts for API clients, caching,
// and parameter generation.
package port

import "chief-checker/internal/service/checkers/checkerModels/requestModels"

// ParamGenerator defines the interface for generating request parameters.
// It provides functionality for creating signed request parameters and nonce values
// required for secure API communication.
type ParamGenerator interface {
	// Generate creates request parameters including signature and nonce.
	//
	// Parameters:
	// - payload: request parameters to sign
	// - method: HTTP method for the request
	// - path: endpoint path
	//
	// Returns:
	// - *requestModels.RequestParams: generated parameters including signature
	// - error: if parameter generation fails
	Generate(payload map[string]string, method, path string) (*requestModels.RequestParams, error)
}

// IDGenerate defines the interface for generating unique identifiers.
// It provides functionality for creating random identifiers of specified lengths.
type IDGenerate interface {
	// Generate creates an identifier of the specified length.
	//
	// Parameters:
	// - length: desired length of the identifier
	//
	// Returns:
	// - string: generated identifier
	Generate(length int) string
}
