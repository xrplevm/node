package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterInvariants registers all module invariants
func RegisterInvariants(_ sdk.InvariantRegistry, _ Keeper) {}
