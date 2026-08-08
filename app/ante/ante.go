package ante

import (
	baseevmante "github.com/cosmos/evm/ante"

	errorsmod "cosmossdk.io/errors"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
)

type anteHandlerFactory func(sdk.Context, baseevmante.HandlerOptions) sdk.AnteHandler

// NewAnteHandler returns an ante handler responsible for attempting to route an
// Ethereum or SDK transaction to an internal ante handler for performing
// transaction-level processing (e.g. fee payment, signature verification) before
// being passed onto it's respective handler.
func NewAnteHandler(options baseevmante.HandlerOptions) sdk.AnteHandler {
	return newAnteHandler(options, newCosmosAnteHandler, newMonoEVMAnteHandler)
}

func newAnteHandler(
	options baseevmante.HandlerOptions,
	cosmosHandler anteHandlerFactory,
	evmHandler anteHandlerFactory,
) sdk.AnteHandler {
	return func(
		ctx sdk.Context, tx sdk.Tx, sim bool,
	) (newCtx sdk.Context, err error) {
		if tx == nil {
			return ctx, errorsmod.Wrap(errortypes.ErrUnknownRequest, "tx is nil")
		}

		var opts []*codectypes.Any
		if txExt, ok := tx.(ante.HasExtensionOptionsTx); ok {
			opts = txExt.GetExtensionOptions()
		}
		if len(opts) == 0 {
			return cosmosHandler(ctx, options)(ctx, tx, sim)
		}

		var anteHandler sdk.AnteHandler
		switch typeURL := opts[0].GetTypeUrl(); typeURL {
		case "/cosmos.evm.vm.v1.ExtensionOptionsEthereumTx":
			// handle as *evmtypes.MsgEthereumTx
			anteHandler = evmHandler(ctx, options)
		case "/cosmos.evm.ante.v1.ExtensionOptionDynamicFeeTx":
			// cosmos-sdk tx with dynamic fee extension
			anteHandler = cosmosHandler(ctx, options)
		default:
			return ctx, errorsmod.Wrapf(
				errortypes.ErrUnknownExtensionOptions,
				"rejecting tx with unsupported extension option: %s", typeURL,
			)
		}
		return anteHandler(ctx, tx, sim)
	}
}
