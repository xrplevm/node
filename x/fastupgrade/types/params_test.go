package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/xrplevm/node/v12/x/fastupgrade/types"
)

func TestDefaultParams(t *testing.T) {
	defaultParams := types.DefaultParams()
	require.False(t, defaultParams.Enabled)
	require.NoError(t, types.DefaultGenesis().Validate())
}

func TestParams_Validate(t *testing.T) {
	tt := []struct {
		name        string
		mutate      func(p *types.Params)
		expectedErr error
	}{
		{
			name:   "should pass - timelock at the minimum",
			mutate: func(p *types.Params) { p.TimelockBlocks = 2 },
		},
		{
			name:        "should fail - timelock below the minimum",
			mutate:      func(p *types.Params) { p.TimelockBlocks = 1 },
			expectedErr: types.ErrInvalidParams,
		},
		{
			name:   "should pass - timelock at the maximum",
			mutate: func(p *types.Params) { p.TimelockBlocks = 100000 },
		},
		{
			name:        "should fail - timelock above the maximum",
			mutate:      func(p *types.Params) { p.TimelockBlocks = 100001 },
			expectedErr: types.ErrInvalidParams,
		},
		{
			name:   "should pass - postpone at the minimum",
			mutate: func(p *types.Params) { p.PostponeBlocks = 1 },
		},
		{
			name:        "should fail - postpone below the minimum",
			mutate:      func(p *types.Params) { p.PostponeBlocks = 0 },
			expectedErr: types.ErrInvalidParams,
		},
		{
			name:   "should pass - postpone at the maximum",
			mutate: func(p *types.Params) { p.PostponeBlocks = 100000 },
		},
		{
			name:        "should fail - postpone above the maximum",
			mutate:      func(p *types.Params) { p.PostponeBlocks = 100001 },
			expectedErr: types.ErrInvalidParams,
		},
		{
			name:   "should pass - threshold of one",
			mutate: func(p *types.Params) { p.Threshold = "1" },
		},
		{
			name:   "should pass - smallest threshold",
			mutate: func(p *types.Params) { p.Threshold = "0.000000000000000001" },
		},
		{
			name:        "should fail - threshold of zero",
			mutate:      func(p *types.Params) { p.Threshold = "0" },
			expectedErr: types.ErrInvalidParams,
		},
		{
			name:        "should fail - negative threshold",
			mutate:      func(p *types.Params) { p.Threshold = "-0.5" },
			expectedErr: types.ErrInvalidParams,
		},
		{
			name:        "should fail - threshold above one",
			mutate:      func(p *types.Params) { p.Threshold = "1.000000000000000001" },
			expectedErr: types.ErrInvalidParams,
		},
		{
			name:        "should fail - threshold with 19 decimals",
			mutate:      func(p *types.Params) { p.Threshold = "0.8500000000000000001" },
			expectedErr: types.ErrInvalidParams,
		},
		{
			name:        "should fail - empty threshold",
			mutate:      func(p *types.Params) { p.Threshold = "" },
			expectedErr: types.ErrInvalidParams,
		},
		{
			name:        "should fail - threshold is not a number",
			mutate:      func(p *types.Params) { p.Threshold = "abc" },
			expectedErr: types.ErrInvalidParams,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			params := types.DefaultParams()
			tc.mutate(&params)

			err := params.Validate()
			if tc.expectedErr != nil {
				require.ErrorIs(t, err, tc.expectedErr)
				require.ErrorIs(t, (&types.MsgUpdateParams{Params: params}).ValidateBasic(), tc.expectedErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
