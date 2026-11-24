package app_test

import (
	"math/big"
	"testing"
	"time"

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

// TestSetOsakaTime verifies that SetOsakaTime writes the expected Osaka
// activation timestamp for each chain ID, keeps the EVM chain config valid,
// and flips IsOsaka from false at t-1s to true at t.
func TestSetOsakaTime(t *testing.T) {
	resetChainConfig(t)
	t.Cleanup(func() { resetChainConfig(t) })

	tt := []struct {
		name       string
		evmChainID uint64
		expected   uint64
	}{
		{
			name:       "devnet",
			evmChainID: 1449900,
			expected:   0,
		},
		{
			name:       "testnet",
			evmChainID: 1449000,
			expected:   uint64(time.Date(2026, time.November, 25, 11, 0, 0, 0, time.UTC).Unix()), //nolint:gosec // G115
		},
		{
			name:       "mainnet",
			evmChainID: 1440000,
			expected:   uint64(time.Date(2026, time.November, 26, 13, 0, 0, 0, time.UTC).Unix()), //nolint:gosec // G115
		},
		{
			name:       "unlisted chain runs from genesis",
			evmChainID: 1449999,
			expected:   0,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			app.SetOsakaTime(tc.evmChainID)

			require.NoError(t, evmtypes.GetChainConfig().Validate())

			ethCfg := evmtypes.GetEthChainConfig()
			require.NotNil(t, ethCfg.OsakaTime)
			require.Equal(t, tc.expected, *ethCfg.OsakaTime)
			// IsOsaka also checks IsLondon(blockHeight), so pass any height >= 0.
			require.True(t, ethCfg.IsOsaka(big.NewInt(1), tc.expected))
			if tc.expected > 0 {
				require.False(t, ethCfg.IsOsaka(big.NewInt(1), tc.expected-1))
			}
		})
	}
}
