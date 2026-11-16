package app_test

import (
	"testing"

	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/stretchr/testify/require"

	"github.com/xrplevm/node/v10/app"
)

// resetChainConfig replaces the process-wide EVM chain config with the default
// one. It keeps the default EVM chain ID, because only that ID lets a later
// test set the config again.
func resetChainConfig(t *testing.T) {
	t.Helper()
	require.NoError(t, evmtypes.SetChainConfig(evmtypes.DefaultChainConfig(evmtypes.DefaultEVMChainID)))
}

func TestDisableOsaka(t *testing.T) {
	resetChainConfig(t)
	t.Cleanup(func() { resetChainConfig(t) })

	// The EVM module enables Osaka by default. If this changes, DisableOsaka
	// is no longer needed.
	require.NotNil(t, evmtypes.GetEthChainConfig().OsakaTime)

	app.DisableOsaka()

	// DisableOsaka writes through the pointer that GetChainConfig returns. A
	// fresh read sees the change, which proves that the pointer is the shared
	// config that the EVM module reads, and not a copy.
	require.Nil(t, evmtypes.GetEthChainConfig().OsakaTime)
}
