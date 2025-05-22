// Package account provides domain entities and operations for managing blockchain accounts.
package account

import (
	"chief-checker/pkg/errors"
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

// Account represents an Ethereum account with its address and optional private key.
// It can be initialized either with just an address or with a private key from which
// the address will be derived.
type Account struct {
	Address    common.Address    // Ethereum address
	PrivateKey *ecdsa.PrivateKey // Optional private key (may be nil for address-only accounts)
}

// AccountFactory creates a slice of accounts from the provided data.
// It supports different initialization types through the walletType parameter.
//
// Parameters:
// - addressData: slice of strings containing either addresses or private keys
// - walletType: type of initialization (with private key or address only)
//
// Returns:
// - []*Account: slice of initialized accounts
// - error: if initialization fails
//
// The function validates all inputs and ensures proper initialization based on the wallet type.
func AccountFactory(addressData []string, walletType AccountDomainType) ([]*Account, error) {
	if len(addressData) == 0 {
		return nil, errors.Wrap(errors.ErrInvalidParams, "empty address data")
	}

	switch walletType {
	case AccountWithPrivateKey:
		return initWithPrivateKey(addressData)
	case AccountWithAddress:
		return initSimpleAddress(addressData)
	default:
		return nil, errors.Wrap(errors.ErrInvalidParams, "invalid module")
	}
}

// initWithPrivateKey creates accounts from private keys.
// Each private key is parsed and used to derive its corresponding address.
//
// Parameters:
// - addressData: slice of private keys in hex format
//
// Returns:
// - []*Account: slice of accounts with both address and private key
// - error: if any private key is invalid or address derivation fails
func initWithPrivateKey(addressData []string) ([]*Account, error) {
	if len(addressData) == 0 {
		return nil, errors.Wrap(errors.ErrInvalidParams, "empty address data")
	}

	accounts := make([]*Account, len(addressData))
	for i, hexPk := range addressData {
		pk, err := ParsePrivateKey(hexPk)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse private key")
		}
		address, err := DeriveAddress(pk)
		if err != nil {
			return nil, errors.Wrap(err, "failed to derive address")
		}
		accounts[i] = &Account{
			Address:    address,
			PrivateKey: pk,
		}
	}
	return accounts, nil
}

// initSimpleAddress creates accounts from either addresses or private keys.
// For each input:
// 1. Tries to parse it as a private key and derive the address
// 2. If that fails, tries to parse it as an Ethereum address
//
// Parameters:
// - addressData: slice of strings containing either addresses or private keys
//
// Returns:
// - []*Account: slice of initialized accounts
// - error: if any input is invalid or no accounts could be created
func initSimpleAddress(addressData []string) ([]*Account, error) {
	if len(addressData) == 0 {
		return nil, errors.Wrap(errors.ErrInvalidParams, "empty address data")
	}

	accounts := make([]*Account, 0, len(addressData))

	for i, line := range addressData {
		pk, err := ParsePrivateKey(line)
		if err == nil {
			address, err := DeriveAddress(pk)
			if err == nil {
				accounts = append(accounts, &Account{
					Address:    address,
					PrivateKey: pk,
				})
				continue
			}
		}

		if err := validateAddressFormat(line); err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("invalid hex format at index %d: %s", i, line))
		}
		accounts = append(accounts, &Account{
			Address: common.HexToAddress(line),
		})
	}

	if len(accounts) == 0 {
		return nil, errors.Wrap(errors.ErrNoCreatedValue, "no valid accounts created")
	}

	return accounts, nil
}
