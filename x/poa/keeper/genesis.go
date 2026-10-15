package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/xrplevm/node/v12/x/poa/types"
)

// ExportGenesis returns the module's exported genesis
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
