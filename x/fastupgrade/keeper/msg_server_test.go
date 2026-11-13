package keeper_test

import (
	"testing"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"

	"github.com/xrplevm/node/v12/x/fastupgrade/keeper"
	"github.com/xrplevm/node/v12/x/fastupgrade/types"
)

func TestMsgServer_UpdateParams(t *testing.T) {
	govAddr := authtypes.NewModuleAddress(govtypes.ModuleName).String()
	otherAddr := authtypes.NewModuleAddress("other").String()

	valid := types.Params{Enabled: true, TimelockBlocks: 2, Threshold: "1", PostponeBlocks: 1}
	incompatible := types.Params{Enabled: true, TimelockBlocks: 150, Threshold: "0.6", PostponeBlocks: 10}
	incompatibleDisabled := incompatible
	incompatibleDisabled.Enabled = false

	tt := []struct {
		name        string
		authority   string
		params      types.Params
		expectedErr error
	}{
		{
			name:        "should fail - authority is not the gov module",
			authority:   otherAddr,
			params:      valid,
			expectedErr: govtypes.ErrInvalidSigner,
		},
		{
			name:        "should fail - invalid params",
			authority:   govAddr,
			params:      types.Params{TimelockBlocks: 1, Threshold: "0.85"},
			expectedErr: types.ErrInvalidParams,
		},
		{
			name:        "should fail - enabled with a threshold incompatible with gov",
			authority:   govAddr,
			params:      incompatible,
			expectedErr: types.ErrIncompatibleWithGov,
		},
		{
			// Gov can always disable the module, also with an incompatible
			// threshold.
			name:      "should pass - disabled with a threshold incompatible with gov",
			authority: govAddr,
			params:    incompatibleDisabled,
		},
		{
			name:      "should pass - valid params",
			authority: govAddr,
			params:    valid,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t)

			_, err := keeper.NewMsgServerImpl(f.k).UpdateParams(f.ctx, &types.MsgUpdateParams{Authority: tc.authority, Params: tc.params})
			if tc.expectedErr != nil {
				require.ErrorIs(t, err, tc.expectedErr)
				return
			}
			require.NoError(t, err)

			res, err := keeper.NewQueryServerImpl(f.k).Params(f.ctx, &types.QueryParamsRequest{})
			require.NoError(t, err)
			require.Equal(t, tc.params, res.Params)
		})
	}
}
