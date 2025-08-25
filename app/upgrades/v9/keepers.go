package v9

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/ethereum/go-ethereum/common"
)

type ERC20Keeper interface {
	SetDynamicPrecompile(ctx sdk.Context, precompile common.Address)
	SetNativePrecompile(ctx sdk.Context, precompile common.Address)
}

type EvmKeeper interface {
	GetParams(ctx sdk.Context) evmtypes.Params
	SetParams(ctx sdk.Context, params evmtypes.Params) error
	SetCodeHash(ctx sdk.Context, addrBytes, hashBytes []byte)
}
