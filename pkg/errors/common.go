// Package errors provides a centralized error handling system.
// It defines common error types and utilities for error wrapping and context preservation.
package errors

import (
	"errors"
	"fmt"
)

var (
	// OS and general errors

	// ErrNotFound indicates that a requested resource was not found.
	ErrNotFound = New("not found")

	// ErrUnsupportedType indicates an operation on an unsupported data type.
	ErrUnsupportedType = New("unsupported type")

	// ErrExitProgram indicates a normal program termination request.
	ErrExitProgram = New("exit program")

	// ErrValueEmpty indicates that a required value is empty or nil.
	ErrValueEmpty = New("value empty")

	// ErrSelection indicates an error in user selection or input.
	ErrSelection = New("selection error")

	// ErrUnexpectedType indicates that a value has an unexpected type.
	ErrUnexpectedType = New("unexpected type")

	// HTTP and network errors

	// ErrConnectionFailed indicates a network connection failure.
	ErrConnectionFailed = New("connection failed")

	// ErrRequestFailed indicates a failed HTTP request.
	ErrRequestFailed = New("request failed")

	// ErrResponseParsing indicates failure to parse an HTTP response.
	ErrResponseParsing = New("failed to parse response")

	// ErrStatusCode indicates an unexpected HTTP status code.
	ErrStatusCode = New("unexpected status code")

	// ErrRateLimitReached indicates that an API rate limit has been reached.
	ErrRateLimitReached = New("rate limit reached")

	// Service errors

	// ErrInvalidParams indicates invalid or missing parameters.
	ErrInvalidParams = New("invalid or missing parameters")

	// ErrNoCreatedValue indicates failure to create a required value.
	ErrNoCreatedValue = New("no parameter has been created")

	// ErrFailedInit indicates initialization failure.
	ErrFailedInit = New("failed to initialize")

	// Blockchain and transaction errors

	// ErrGasEstimation indicates failure to estimate transaction gas.
	ErrGasEstimation = New("failed to estimate gas")

	// ErrTransactionFailed indicates a failed blockchain transaction.
	ErrTransactionFailed = New("transaction failed")

	// ErrTransactionTimeout indicates a transaction timeout.
	ErrTransactionTimeout = New("transaction wait timeout")

	// ErrContractCall indicates a smart contract call failure.
	ErrContractCall = New("contract call failed")

	// ErrChainID indicates failure to get blockchain chain ID.
	ErrChainID = New("failed to get chain ID")

	// ErrNonceRetrieval indicates failure to get transaction nonce.
	ErrNonceRetrieval = New("failed to get nonce")

	// ErrTxSigning indicates transaction signing failure.
	ErrTxSigning = New("failed to sign transaction")

	// ErrTxSending indicates transaction broadcast failure.
	ErrTxSending = New("failed to send transaction")

	// ErrBalanceEstimation indicates failure to estimate balance.
	ErrBalanceEstimation = New("failed to estimate balance")

	// Wallet and key management errors

	// ErrEntropyGeneration indicates failure to generate entropy.
	ErrEntropyGeneration = New("failed to generate entropy")

	// ErrMnemonicGeneration indicates failure to generate mnemonic.
	ErrMnemonicGeneration = New("failed to generate mnemonic")

	// ErrKeyDerivation indicates key derivation failure.
	ErrKeyDerivation = New("failed to derive key")

	// ErrAddressGeneration indicates address generation failure.
	ErrAddressGeneration = New("failed to generate address")

	// ErrInvalidKeyFormat indicates invalid key format.
	ErrInvalidKeyFormat = New("invalid key format")

	// ErrInvalidSeedSize indicates invalid seed size.
	ErrInvalidSeedSize = New("invalid seed size")

	// Bridge errors

	// ErrBridgeValidation indicates invalid bridge input.
	ErrBridgeValidation = New("invalid bridge input")

	// ErrQuoteRetrieval indicates failure to get price quote.
	ErrQuoteRetrieval = New("failed to get quote")

	// ErrInvalidQuote indicates invalid quote response.
	ErrInvalidQuote = New("invalid quote response")

	// ErrHighImpact indicates unacceptable price impact.
	ErrHighImpact = New("price impact too high")

	// CEX errors

	// ErrCexFailed indicates a failed CEX operation.
	ErrCexFailed = New("invalid request")

	// ErrTokenNotFound indicates that a token was not found.
	ErrTokenNotFound = New("token not found")

	// ErrPriceEstimation indicates failure to estimate price.
	ErrPriceEstimation = New("failed to estimate price")

	// Config errors

	// ErrConfigRead indicates failed to read config.
	ErrConfigRead = New("failed to read config")

	// ErrConfigParse indicates failed to parse config.
	ErrConfigParse = New("failed to parse config")

	// ErrConfigSave indicates failed to save config.
	ErrConfigSave = New("failed to save config")

	// ErrInvalidConfig indicates invalid configuration.
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
