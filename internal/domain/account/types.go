// Package account provides domain entities and operations for managing blockchain accounts.
package account

// AccountDomainType represents the type of account initialization.
// It determines how the account will be created and what data is required.
type AccountDomainType string

const (
	// AccountWithPrivateKey indicates that the account should be initialized with a private key.
	// This type requires a valid private key in hex format and will derive the address from it.
	AccountWithPrivateKey AccountDomainType = "With Private Key"

	// AccountWithAddress indicates that the account should be initialized with just an address.
	// This type accepts either a private key (which will be used to derive the address)
	// or a plain Ethereum address in hex format.
	AccountWithAddress AccountDomainType = "With Address"
)
