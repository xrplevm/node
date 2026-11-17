package app

import (
	evmtypes "github.com/cosmos/evm/x/vm/types"
)

// DisableOsaka turns the Osaka hard fork off in the process-wide EVM chain
// config, on every network. The EVM module enables Osaka from genesis by
// default, and the config is not stored in chain state, so the node must
// override it on each start. Call it after evmkeeper.NewKeeper, which creates
// the config.
func DisableOsaka() {
	// GetChainConfig returns the pointer that the EVM module reads on each use.
	evmtypes.GetChainConfig().OsakaTime = nil
}
