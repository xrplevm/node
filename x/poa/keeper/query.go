package keeper

import (
	"github.com/node/xrplevm/v2/x/poa/types"
)

var _ types.QueryServer = Keeper{}
