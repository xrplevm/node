package keeper

import (
	"github.com/xrplevm/node/v9/x/poa/types"
)

var _ types.QueryServer = Keeper{}
