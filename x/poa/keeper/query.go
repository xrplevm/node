package keeper

import (
	"github.com/xrplevm/node/v2/x/poa/types"
)

var _ types.QueryServer = Keeper{}
