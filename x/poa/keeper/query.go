package keeper

import (
	"github.com/Peersyst/exrp/v2/x/poa/types"
)

var _ types.QueryServer = Keeper{}
