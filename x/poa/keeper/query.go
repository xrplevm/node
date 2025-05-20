package keeper

import (
	"github.com/xrplevm/node/v8/x/poa/types"
)

var _ types.QueryServer = Keeper{}
