package types_test

import (
	"fmt"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/xrplevm/node/v12/x/fastupgrade/types"
)

func TestDefaultParams(t *testing.T) {
	defaultParams := types.DefaultParams()
	require.False(t, defaultParams.Enabled)
	require.Empty(t, defaultParams.Proposers)
	require.NoError(t, types.DefaultGenesis().Validate())
}

// These tests do not set the chain's address prefix, so the addresses have
// the SDK's default prefix.
var (
	proposerA = sdk.AccAddress("proposer-a__________").String()
	proposerB = sdk.AccAddress("proposer-b__________").String()
)

// proposers returns n different addresses.
func proposers(n int) []string {
	list := make([]string, n)
	for i := range list {
		list[i] = sdk.AccAddress(fmt.Sprintf("proposer-%02d_________", i)).String()
	}
	return list
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
		{
			name:   "should pass - two proposers",
			mutate: func(p *types.Params) { p.Proposers = []string{proposerA, proposerB} },
		},
		{
			// Only the decode error rejects this entry: a failed decode gives
			// an empty address, which prints as "".
			name:        "should fail - empty proposer",
			mutate:      func(p *types.Params) { p.Proposers = []string{""} },
			expectedErr: types.ErrInvalidParams,
		},
		{
			// Gov stores the proposer in lower case, so this entry would
			// never match.
			name:        "should fail - proposer in upper case",
			mutate:      func(p *types.Params) { p.Proposers = []string{strings.ToUpper(proposerA)} },
			expectedErr: types.ErrInvalidParams,
		},
		{
			name:        "should fail - duplicate proposer",
			mutate:      func(p *types.Params) { p.Proposers = []string{proposerA, proposerB, proposerA} },
			expectedErr: types.ErrInvalidParams,
		},
		{
			name:   "should pass - proposers at the maximum",
			mutate: func(p *types.Params) { p.Proposers = proposers(16) },
		},
		{
			name:        "should fail - proposers above the maximum",
			mutate:      func(p *types.Params) { p.Proposers = proposers(17) },
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
