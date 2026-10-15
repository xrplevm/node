package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/xrplevm/node/v12/x/poa/types"
)

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(_ sdk.Context) types.Params {
	return types.NewParams()
}
