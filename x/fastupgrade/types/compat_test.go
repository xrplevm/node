package types_test

import (
	"testing"

	"cosmossdk.io/math"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/stretchr/testify/require"

	"github.com/xrplevm/node/v12/x/fastupgrade/types"
)

func liveGovParams() govv1.Params {
	p := govv1.DefaultParams()
	p.Quorum = "0.667000000000000000"
	p.Threshold = "0.667000000000000000"
	p.VetoThreshold = "0.334000000000000000"
	p.ExpeditedThreshold = "0.800000000000000000"
	return p
}

func TestCheckGovCompatibility(t *testing.T) {
	tt := []struct {
		name      string
		threshold string
		// mutate changes the live gov params. It is optional.
		mutate      func(p *govv1.Params)
		expectedErr error
	}{
		{
			name:      "should pass - default threshold with the live gov params",
			threshold: types.DefaultThreshold.String(),
		},
		{
			name:      "should pass - default threshold with the SDK default gov params",
			threshold: types.DefaultThreshold.String(),
			mutate:    func(p *govv1.Params) { *p = govv1.DefaultParams() },
		},
		{
			name:      "should fail - equal to the gov threshold",
			threshold: "0.7",
			mutate: func(p *govv1.Params) {
				p.Threshold = "0.7"
				p.ExpeditedThreshold = "0.69"
			},
			expectedErr: types.ErrIncompatibleWithGov,
		},
		{
			name:      "should pass - one unit above the gov threshold",
			threshold: "0.700000000000000001",
			mutate: func(p *govv1.Params) {
				p.Threshold = "0.7"
				p.ExpeditedThreshold = "0.69"
			},
		},
		{
			name:        "should fail - equal to the gov expedited threshold",
			threshold:   "0.8",
			expectedErr: types.ErrIncompatibleWithGov,
		},
		{
			name:      "should pass - one unit above the gov expedited threshold",
			threshold: "0.800000000000000001",
		},
		{
			name:      "should pass - equal to the gov quorum",
			threshold: "0.9",
			mutate:    func(p *govv1.Params) { p.Quorum = "0.9" },
		},
		{
			name:        "should fail - one unit below the gov quorum",
			threshold:   "0.899999999999999999",
			mutate:      func(p *govv1.Params) { p.Quorum = "0.9" },
			expectedErr: types.ErrIncompatibleWithGov,
		},
		{
			name:      "should pass - equal to one minus the gov veto threshold",
			threshold: "0.9",
			mutate:    func(p *govv1.Params) { p.VetoThreshold = "0.1" },
		},
		{
			name:        "should fail - one unit below one minus the gov veto threshold",
			threshold:   "0.899999999999999999",
			mutate:      func(p *govv1.Params) { p.VetoThreshold = "0.1" },
			expectedErr: types.ErrIncompatibleWithGov,
		},
		{
			name:        "should fail - gov param is not a number",
			threshold:   "0.85",
			mutate:      func(p *govv1.Params) { p.Quorum = "x" },
			expectedErr: types.ErrIncompatibleWithGov,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			govParams := liveGovParams()
			if tc.mutate != nil {
				tc.mutate(&govParams)
			}

			err := types.CheckGovCompatibility(math.LegacyMustNewDecFromStr(tc.threshold), govParams)
			if tc.expectedErr != nil {
				require.ErrorIs(t, err, tc.expectedErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
