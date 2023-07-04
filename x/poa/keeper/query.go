package keeper

import (
	"github.com/Peersyst/exrp/x/poa/types"
)

var _ types.QueryServer = Keeper{}
