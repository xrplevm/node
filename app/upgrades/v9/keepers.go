package v9

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
)

type ERC20Keeper interface {
	SetDynamicPrecompile(ctx sdk.Context, precompile common.Address)
	SetNativePrecompile(ctx sdk.Context, precompile common.Address)
}
