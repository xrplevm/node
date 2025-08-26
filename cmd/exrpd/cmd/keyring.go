package cmd

import (
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cosmosLedger "github.com/cosmos/cosmos-sdk/crypto/ledger"
	evmhd "github.com/cosmos/evm/crypto/hd"
	evmkeyring "github.com/cosmos/evm/crypto/keyring"
	legacyhd "github.com/xrplevm/node/v9/legacy/crypto/hd"
)

var (
	SupportedAlgorithms       = keyring.SigningAlgoList{evmhd.EthSecp256k1, legacyhd.EthSecp256k1}
	SupportedAlgorithmsLedger = keyring.SigningAlgoList{evmhd.EthSecp256k1, legacyhd.EthSecp256k1}
)

func CustomKeyringOption() keyring.Option {
	return func(options *keyring.Options) {
		options.SupportedAlgos = SupportedAlgorithms
		options.SupportedAlgosLedger = SupportedAlgorithmsLedger
		options.LedgerDerivation = func() (cosmosLedger.SECP256K1, error) { return evmkeyring.LedgerDerivation() }
		options.LedgerCreateKey = evmkeyring.CreatePubkey
		options.LedgerAppName = evmkeyring.AppName
		options.LedgerSigSkipDERConv = evmkeyring.SkipDERConversion

	}
}
