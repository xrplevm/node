package keeper

import (
	"github.com/xrplevm/node/v3/x/poa/types"
)

var _ types.QueryServer = Keeper{}
