package keeper

import (
	"github.com/xrplevm/node/v7/x/poa/types"
)

var _ types.QueryServer = Keeper{}
