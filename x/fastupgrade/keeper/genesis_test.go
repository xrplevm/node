package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/xrplevm/node/v12/x/fastupgrade/types"
)

// The export has only the params, also while a timer is running.
func TestKeeper_ExportGenesis(t *testing.T) {
	f := newFixture(t)
	f.startTimer(f.proposeMinor())

	gs, err := f.k.ExportGenesis(f.ctx)
	require.NoError(t, err)
	require.Equal(t, &types.GenesisState{Params: types.Params{Enabled: true, TimelockBlocks: 150, Threshold: "0.850000000000000000", PostponeBlocks: 10}}, gs)
}

func TestKeeper_InitGenesis(t *testing.T) {
	tt := []struct {
		name        string
		params      types.Params
		expectedErr error
	}{
		{
			name:        "should fail - invalid params",
			params:      types.Params{TimelockBlocks: 0, Threshold: "0.85"},
			expectedErr: types.ErrInvalidParams,
		},
		{
			// Gov compatibility is a runtime check, so an export taken after
			// gov params drifted still imports.
			name:   "should pass - enabled with a threshold incompatible with gov",
			params: types.Params{Enabled: true, TimelockBlocks: 150, Threshold: "0.5", PostponeBlocks: 10},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t)

			err := f.k.InitGenesis(f.ctx, types.GenesisState{Params: tc.params})
			if tc.expectedErr != nil {
				require.ErrorIs(t, err, tc.expectedErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
