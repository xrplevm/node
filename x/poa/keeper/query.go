package keeper

import (
	"github.com/xrplevm/node/v5/x/poa/types"
)

var _ types.QueryServer = Keeper{}
