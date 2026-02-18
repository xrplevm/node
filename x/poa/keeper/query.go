package keeper

import (
	"github.com/xrplevm/node/v10/x/poa/types"
)

var _ types.QueryServer = Keeper{}
