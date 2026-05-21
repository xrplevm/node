package v11

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	icahosttypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/host/types"
)

// ICAHostKeeper is the narrow interface required by the v11 upgrade
// handler. It matches a subset of icahostkeeper.Keeper.
type ICAHostKeeper interface {
	SetParams(ctx sdk.Context, params icahosttypes.Params)
}
