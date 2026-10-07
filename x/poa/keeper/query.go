package keeper

import (
	"github.com/xrplevm/node/v12/x/poa/types"
)

var _ types.QueryServer = Keeper{}
