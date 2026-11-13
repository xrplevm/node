package v12

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	evmtypes "github.com/cosmos/evm/x/vm/types"

	fastupgradetypes "github.com/xrplevm/node/v12/x/fastupgrade/types"
)

// EvmKeeper is the narrow interface required by the v12 upgrade handler.
type EvmKeeper interface {
	GetParams(ctx sdk.Context) evmtypes.Params
	SetParams(ctx sdk.Context, params evmtypes.Params) error
}

// FastUpgradeKeeper is the narrow interface required by the v12 upgrade handler.
type FastUpgradeKeeper interface {
	SetParams(ctx sdk.Context, params fastupgradetypes.Params) error
}
