package keeper

import (
	"github.com/xrplevm/node/v6/x/poa/types"
)

var _ types.QueryServer = Keeper{}
