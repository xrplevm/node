package app

import (
	sdkmath "cosmossdk.io/math"
	evmtypes "github.com/cosmos/evm/x/vm/types"
)

const (
	mainnetEVMChainID uint64 = 1440000
	testnetEVMChainID uint64 = 1449000
	devnetEVMChainID  uint64 = 1449900
)

// osakaTimeByChainID maps the EVM chain ID of each existing network to its
// Osaka activation time, in Unix seconds. Unlisted networks run Osaka from
// genesis (timestamp 0).
var osakaTimeByChainID = map[uint64]uint64{
	devnetEVMChainID:  0,
	testnetEVMChainID: 1795604400, // 2026-11-25 11:00:00 UTC
	mainnetEVMChainID: 1795698000, // 2026-11-26 13:00:00 UTC
}

// SetOsakaTime sets the Osaka activation time in the process-wide EVM chain
// config for the given EVM chain ID. Unlisted chains run Osaka from genesis.
// Call it after evmkeeper.NewKeeper, which creates the config.
func SetOsakaTime(evmChainID uint64) {
	// GetChainConfig returns the pointer that the EVM module reads on each use.
	osakaTime := sdkmath.NewIntFromUint64(osakaTimeByChainID[evmChainID])
	evmtypes.GetChainConfig().OsakaTime = &osakaTime
}
