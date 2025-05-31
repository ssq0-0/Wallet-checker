// Package serverInterface defines the core interfaces for the server package.
// It provides contracts for server components, handlers, and data formatting.
package serverInterface

import (
	"chief-checker/internal/service/server/serverTypes"
	"chief-checker/internal/usecase/features/checkerUsecase/types"
	"net/http"
)

// Handler defines the interface for HTTP request handlers.
// It provides methods for handling balance and address-related requests.
type Handler interface {
	// BalanceData handles requests for balance information.
	// It processes and returns the balance data for the requested account.
	BalanceData(w http.ResponseWriter, r *http.Request)

	// Addresses handles requests for address information.
	// It processes and returns the address data for the requested account.
	Addresses(w http.ResponseWriter, r *http.Request)
}

// AccountStats defines the interface for account statistics.
// It provides methods to retrieve aggregated and global statistics.
type AccountStats interface {
	// GetAllStats returns a map of all account statistics.
	// The key is the account address, and the value is the aggregated data.
	GetAllStats() map[string]*types.AggregatedData

	// GetGlobalStats returns the global statistics across all accounts.
	GetGlobalStats() *types.GlobalStats
}

// Formater defines the interface for data formatting.
// It provides methods to format balance and address data for API responses.
type Formater interface {
	// BalanceDataFormat formats the global statistics into a structured response.
	// It processes token and chain statistics to create a formatted response.
	BalanceDataFormat(stats *types.GlobalStats) *serverTypes.TotalStats

	// AddressDataFormat formats the address data into a structured response.
	// It processes individual address data to create a formatted response.
	AddressDataFormat(data map[string]*types.AggregatedData) []serverTypes.AddressResponse
}

// Server defines the interface for the HTTP server.
// It provides methods to control the server lifecycle and handle requests.
type Server interface {
	// StartServer starts the HTTP server on the specified port.
	// It initializes the server and begins listening for requests.
	StartServer(port string)

	// StopServer handles the server shutdown request.
	// It gracefully stops the server and returns a status response.
	StopServer(w http.ResponseWriter, r *http.Request)

	// Done returns a channel that is closed when the server is stopped.
	// It can be used to wait for server shutdown.
	Done() <-chan struct{}
}
