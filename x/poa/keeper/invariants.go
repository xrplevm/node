package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterInvariants registers all module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {}
