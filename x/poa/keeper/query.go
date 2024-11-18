package keeper

import (
	"github.com/xrplevm/node/v4/x/poa/types"
)

var _ types.QueryServer = Keeper{}
