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

// Constants for private key and address validation
const (
	privateKeyLengthHex   = 64               // Length of private key in hex without 0x prefix
	privateKeyLengthHexOx = 66               // Length of private key in hex with 0x prefix
	privateKeyLengthBytes = 32               // Length of private key in bytes
	hexPrefix             = "0x"             // Standard hex prefix
	hexRegexPattern       = "^[0-9a-fA-F]+$" // Pattern for valid hex characters
	addressLengthHex      = 40               // Length of Ethereum address in hex without 0x prefix
	addressLengthHexOx    = 42               // Length of Ethereum address in hex with 0x prefix
)

// ParsePrivateKey converts a hex string into an ECDSA private key.
// The input can be with or without "0x" prefix.
//
// Parameters:
// - hexKey: private key in hex format
//
// Returns:
// - *ecdsa.PrivateKey: parsed private key
// - error: if the key format is invalid or parsing fails
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

// DeriveAddress derives an Ethereum address from a private key.
// It uses the public key associated with the private key to generate the address.
//
// Parameters:
// - privateKey: ECDSA private key
//
// Returns:
// - common.Address: derived Ethereum address
// - error: if key derivation fails
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

// validateKeyLength checks if the private key hex string has the correct length.
// It accepts keys both with and without "0x" prefix.
func validateKeyLength(hexKey string) error {
	if len(hexKey) != privateKeyLengthHex && len(hexKey) != privateKeyLengthHexOx {
		return errors.ErrInvalidKeyFormat
	}
	return nil
}

// removeHexPrefix removes the "0x" prefix from a hex string if present.
func removeHexPrefix(hexKey string) string {
	if len(hexKey) > 2 && hexKey[:2] == hexPrefix {
		return hexKey[2:]
	}
	return hexKey
}

// validateHexFormat checks if a string contains only valid hexadecimal characters.
func validateHexFormat(hexKey string) error {
	isHex := regexp.MustCompile(hexRegexPattern).MatchString
	if !isHex(strings.ToLower(hexKey)) {
		return errors.ErrInvalidKeyFormat
	}
	return nil
}

// validateAddressFormat checks if a string is a valid Ethereum address format.
// It validates both the length and character set of the address.
//
// Parameters:
// - address: Ethereum address in hex format (with or without 0x prefix)
//
// Returns:
// - error: if the address format is invalid
func validateAddressFormat(address string) error {
	// Check length with 0x prefix
	if len(address) == addressLengthHexOx {
		if !strings.HasPrefix(address, hexPrefix) {
			return errors.ErrInvalidKeyFormat
		}
		address = address[2:]
	} else if len(address) != addressLengthHex {
		return errors.ErrInvalidKeyFormat
	}

	// Verify hex characters only
	isHex := regexp.MustCompile(hexRegexPattern).MatchString
	if !isHex(strings.ToLower(address)) {
		return errors.ErrInvalidKeyFormat
	}
	return nil
}

// hexToBytes converts a hex string to bytes and validates the length.
//
// Parameters:
// - hexKey: hex string to convert
//
// Returns:
// - []byte: converted bytes
// - error: if conversion fails or length is invalid
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

// createECDSAKey creates an ECDSA private key from bytes.
//
// Parameters:
// - privateKeyBytes: raw private key bytes
//
// Returns:
// - *ecdsa.PrivateKey: created private key
// - error: if key creation fails
func createECDSAKey(privateKeyBytes []byte) (*ecdsa.PrivateKey, error) {
	privateKey, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInvalidKeyFormat, err.Error())
	}
	return privateKey, nil
}
