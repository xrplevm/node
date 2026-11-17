package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/xrplevm/node/v12/x/fastupgrade/types"
)

// InitGenesis stores the genesis params. Gov compatibility is not checked
// here: it is checked before every tally, so an export always imports.
func (k Keeper) InitGenesis(ctx sdk.Context, gs types.GenesisState) error {
	if err := gs.Validate(); err != nil {
		return err
	}
	return k.Params.Set(ctx, gs.Params)
}

// ExportGenesis exports only params. The timer is not exported: a chain
// started from an export has no completed-upgrade baseline.
func (k Keeper) ExportGenesis(ctx sdk.Context) (*types.GenesisState, error) {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	return &types.GenesisState{Params: params}, nil
}
