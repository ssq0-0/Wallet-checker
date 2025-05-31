// Package serverHandler provides HTTP request handlers for the server package.
// It implements the Handler interface to process balance and address-related requests.
package serverHandler

import (
	"chief-checker/internal/service/server/serverInterface"
	"encoding/json"
	"net/http"
)

// serverHandler implements the serverInterface.Handler interface.
// It processes HTTP requests and formats responses using the provided formatter.
type serverHandler struct {
	accountStats serverInterface.AccountStats
	formater     serverInterface.Formater
}

// NewServerHandler creates a new instance of the server handler.
// It takes an AccountStats implementation for data retrieval and a Formater implementation for response formatting.
// Returns an implementation of the serverInterface.Handler interface.
func NewServerHandler(account serverInterface.AccountStats, formater serverInterface.Formater) serverInterface.Handler {
	return &serverHandler{accountStats: account, formater: formater}
}

// BalanceData handles requests for balance information.
// It validates the request method, retrieves global statistics,
// formats the data using the formatter, and sends the response.
// Only GET requests are allowed.
func (h *serverHandler) BalanceData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tokenData := h.formater.BalanceDataFormat(h.accountStats.GetGlobalStats())
	h.sendResponse(w, tokenData)
}

// Addresses handles requests for address information.
// It validates the request method, retrieves all account statistics,
// formats the data using the formatter, and sends the response.
// Only GET requests are allowed.
func (h *serverHandler) Addresses(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	accountData := h.formater.AddressDataFormat(h.accountStats.GetAllStats())

	h.sendResponse(w, accountData)
}

// sendResponse sends a JSON response to the client.
// It sets the appropriate headers and encodes the provided data as JSON.
// If encoding fails, it sends an internal server error response.
func (h *serverHandler) sendResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
