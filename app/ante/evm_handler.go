package ante

import (
	baseevmante "github.com/cosmos/evm/ante"
	evmante "github.com/cosmos/evm/ante/evm"
	evmtypes "github.com/cosmos/evm/x/vm/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
)

func isEthereumTxExtensionOption(anyType *codectypes.Any) bool {
	_, ok := anyType.GetCachedValue().(*evmtypes.ExtensionOptionsEthereumTx)
	return ok
}

// newMonoEVMAnteHandler creates the sdk.AnteHandler implementation for the EVM transactions.
func newMonoEVMAnteHandler(ctx sdk.Context, options baseevmante.HandlerOptions) sdk.AnteHandler {
	evmParams := options.EvmKeeper.GetParams(ctx)
	feemarketParams := options.FeeMarketKeeper.GetParams(ctx)
	decorators := []sdk.AnteDecorator{
		ante.NewExtensionOptionsDecorator(isEthereumTxExtensionOption),
		evmante.NewEVMMonoDecorator(
			options.AccountKeeper,
			options.FeeMarketKeeper,
			options.EvmKeeper,
			options.MaxTxGasWanted,
			&evmParams,
			&feemarketParams,
		),
		baseevmante.NewTxListenerDecorator(options.PendingTxListener),
	}

	return sdk.ChainAnteDecorators(decorators...)
}
