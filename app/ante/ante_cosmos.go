package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcante "github.com/cosmos/ibc-go/v10/modules/core/ante"

	cosmosante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	evmante "github.com/cosmos/evm/ante/evm"
)

// newCosmosAnteHandler creates the default SDK ante handler for Cosmos transactions
func newCosmosAnteHandler(options HandlerOptions) sdk.AnteHandler {
	return sdk.ChainAnteDecorators(
		cosmosante.NewSetUpContextDecorator(),
		cosmosante.NewExtensionOptionsDecorator(options.ExtensionOptionChecker),
		cosmosante.NewValidateBasicDecorator(),
		cosmosante.NewTxTimeoutHeightDecorator(),
		cosmosante.NewValidateMemoDecorator(options.AccountKeeper),
		cosmosante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		cosmosante.NewDeductFeeDecorator(
			options.AccountKeeper,
			options.BankKeeper,
			options.FeegrantKeeper,
			options.TxFeeChecker,
		),
		cosmosante.NewSetPubKeyDecorator(options.AccountKeeper),
		cosmosante.NewValidateSigCountDecorator(options.AccountKeeper),
		cosmosante.NewSigGasConsumeDecorator(options.AccountKeeper, options.SigGasConsumer),
		cosmosante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		cosmosante.NewIncrementSequenceDecorator(options.AccountKeeper),
		ibcante.NewRedundantRelayDecorator(options.IBCKeeper),
		evmante.NewGasWantedDecorator(options.EvmKeeper, options.FeeMarketKeeper),
	)
}
