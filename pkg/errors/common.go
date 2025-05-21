package errors

import (
	"errors"
	"fmt"
)

var (
	// OS error
	ErrNotFound        = New("not found")
	ErrUnsupportedType = New("unsupported type")
	ErrExitProgram     = New("exit program")
	ErrValueEmpty      = New("value empty")
	ErrSelection       = New("selection error")
	ErrUnexpectedType  = New("unexpected type")

	// HTTP error
	ErrConnectionFailed = New("connection failed")
	ErrRequestFailed    = New("request failed")
	ErrResponseParsing  = New("failed to parse response")
	ErrStatusCode       = New("unexpected status code")
	ErrRateLimitReached = New("rate limit reached")

	// Service error
	ErrInvalidParams  = New("invalid or missing parameters")
	ErrNoCreatedValue = New("no parameter has been created")
	ErrFailedInit     = New("failed to initialize")

	// Contract/Transaction errors
	ErrGasEstimation      = New("failed to estimate gas")
	ErrTransactionFailed  = New("transaction failed")
	ErrTransactionTimeout = New("transaction wait timeout")
	ErrContractCall       = New("contract call failed")
	ErrChainID            = New("failed to get chain ID")
	ErrNonceRetrieval     = New("failed to get nonce")
	ErrTxSigning          = New("failed to sign transaction")
	ErrTxSending          = New("failed to send transaction")
	ErrBalanceEstimation  = New("failed to estimate balance")

	// Wallet Generator errors
	ErrEntropyGeneration  = New("failed to generate entropy")
	ErrMnemonicGeneration = New("failed to generate mnemonic")
	ErrKeyDerivation      = New("failed to derive key")
	ErrAddressGeneration  = New("failed to generate address")
	ErrInvalidKeyFormat   = New("invalid key format")
	ErrInvalidSeedSize    = New("invalid seed size")

	// Bridge errors
	ErrBridgeValidation = New("invalid bridge input")
	ErrQuoteRetrieval   = New("failed to get quote")
	ErrInvalidQuote     = New("invalid quote response")
	ErrHighImpact       = New("price impact too high")

	// CEX errors
	ErrCexFailed       = New("invalid request")
	ErrTokenNotFound   = New("token not found")
	ErrPriceEstimation = New("failed to estimate price")
	// Config errors
	ErrConfigRead    = New("failed to read config")
	ErrConfigParse   = New("failed to parse config")
	ErrConfigSave    = New("failed to save config")
	ErrInvalidConfig = New("invalid configuration")
)

func New(text string) error {
	return errors.New(text)
}

func Wrap(err error, context string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", context, err)
}
