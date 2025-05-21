package account

import (
	"chief-checker/pkg/errors"
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

type Account struct {
	Address    common.Address
	PrivateKey *ecdsa.PrivateKey
}

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
