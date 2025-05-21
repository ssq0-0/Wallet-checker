package account

import (
	"chief-checker/pkg/errors"
	"crypto/ecdsa"
	"encoding/hex"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	privateKeyLengthHex   = 64
	privateKeyLengthHexOx = 66
	privateKeyLengthBytes = 32
	hexPrefix             = "0x"
	hexRegexPattern       = "^[0-9a-fA-F]+$"
	addressLengthHex      = 40
	addressLengthHexOx    = 42
)

func ParsePrivateKey(hexKey string) (*ecdsa.PrivateKey, error) {
	if err := validateKeyLength(hexKey); err != nil {
		return nil, err
	}

	hexKey = removeHexPrefix(hexKey)

	if err := validateHexFormat(hexKey); err != nil {
		return nil, err
	}

	privateKeyBytes, err := hexToBytes(hexKey)
	if err != nil {
		return nil, err
	}

	return createECDSAKey(privateKeyBytes)
}

func DeriveAddress(privateKey *ecdsa.PrivateKey) (common.Address, error) {
	if privateKey == nil {
		return common.Address{}, errors.ErrInvalidParams
	}

	publicKey, ok := privateKey.Public().(*ecdsa.PublicKey)
	if !ok {
		return common.Address{}, errors.ErrKeyDerivation
	}

	return crypto.PubkeyToAddress(*publicKey), nil
}

func validateKeyLength(hexKey string) error {
	if len(hexKey) != privateKeyLengthHex && len(hexKey) != privateKeyLengthHexOx {
		return errors.ErrInvalidKeyFormat
	}
	return nil
}

func removeHexPrefix(hexKey string) string {
	if len(hexKey) > 2 && hexKey[:2] == hexPrefix {
		return hexKey[2:]
	}
	return hexKey
}

func validateHexFormat(hexKey string) error {
	isHex := regexp.MustCompile(hexRegexPattern).MatchString
	if !isHex(strings.ToLower(hexKey)) {
		return errors.ErrInvalidKeyFormat
	}
	return nil
}

func validateAddressFormat(address string) error {
	// Проверяем длину с префиксом 0x
	if len(address) == addressLengthHexOx {
		if !strings.HasPrefix(address, hexPrefix) {
			return errors.ErrInvalidKeyFormat
		}
		address = address[2:]
	} else if len(address) != addressLengthHex {
		return errors.ErrInvalidKeyFormat
	}

	// Проверяем, что адрес состоит только из hex символов
	isHex := regexp.MustCompile(hexRegexPattern).MatchString
	if !isHex(strings.ToLower(address)) {
		return errors.ErrInvalidKeyFormat
	}
	return nil
}

func hexToBytes(hexKey string) ([]byte, error) {
	privateKeyBytes, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInvalidKeyFormat, "invalid hex format")
	}

	if len(privateKeyBytes) != privateKeyLengthBytes {
		return nil, errors.ErrInvalidKeyFormat
	}

	return privateKeyBytes, nil
}

func createECDSAKey(privateKeyBytes []byte) (*ecdsa.PrivateKey, error) {
	privateKey, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInvalidKeyFormat, err.Error())
	}
	return privateKey, nil
}
