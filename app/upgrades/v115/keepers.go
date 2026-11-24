package v115

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	evmtypes "github.com/cosmos/evm/x/vm/types"
)

type EvmKeeper interface {
	GetParams(ctx sdk.Context) evmtypes.Params
	SetParams(ctx sdk.Context, params evmtypes.Params) error
}
