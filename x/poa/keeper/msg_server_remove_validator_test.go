package keeper

import (
	"bytes"
	"errors"
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	"github.com/xrplevm/node/v10/x/poa/testutil"
	"github.com/xrplevm/node/v10/x/poa/types"
	"go.uber.org/mock/gomock"
)

func TestMsgServer_RemoveValidator(t *testing.T) {
	ctrl := gomock.NewController(t)

	tt := []struct {
		name             string
		authority        string
		validatorAddress string
		stakingMocks     func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper)
		bankMocks        func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper)
		expectedErr      error
		expectedLog      string
	}{
		{
			name:             "should fail - invalid authority address",
			authority:        "invalidauthority",
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
			stakingMocks:     func(_ sdk.Context, _ *testutil.MockStakingKeeper) {},
			bankMocks:        func(_ sdk.Context, _ *testutil.MockBankKeeper) {},
			expectedErr:      govtypes.ErrInvalidSigner,
		},
		{
			name:             "should fail - invalid validator address",
			authority:        poaAuthority,
			validatorAddress: "invalidvalidatoraddress",
			stakingMocks:     func(_ sdk.Context, _ *testutil.MockStakingKeeper) {},
			bankMocks:        func(_ sdk.Context, _ *testutil.MockBankKeeper) {},
			expectedErr:      errors.New("decoding bech32 failed"),
		},
		{
			name:             "should pass",
			authority:        poaAuthority,
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
					BondDenom: "BND",
				}, nil)
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{
					Tokens: math.NewInt(0),
				}, nil)
				stakingKeeper.EXPECT().GetUnbondingDelegationsFromValidator(ctx, gomock.Any()).Return([]stakingtypes.UnbondingDelegation{}, nil)

				hooks := testutil.NewMockStakingHooks(ctrl)
				hooks.EXPECT().BeforeValidatorModified(ctx, gomock.Any()).Return(nil)
				stakingKeeper.EXPECT().Hooks().Return(hooks)

				stakingKeeper.EXPECT().RemoveValidatorTokens(ctx, gomock.Any(), gomock.Any()).Return(
					stakingtypes.Validator{Status: stakingtypes.Bonded}, nil,
				)
				stakingKeeper.EXPECT().Unbond(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(math.ZeroInt(), nil)
			},
			bankMocks: func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {
				bankKeeper.EXPECT().BurnCoins(ctx, stakingtypes.BondedPoolName, gomock.Any()).Return(nil)
			},
		},
		{
			name:             "should pass - BeforeValidatorModified hook error is swallowed and logged",
			authority:        poaAuthority,
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
			stakingMocks: func(_ sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				// gomock.Any() for ctx because the test swaps the logger after
				// setup, producing a different sdk.Context value.
				stakingKeeper.EXPECT().GetParams(gomock.Any()).Return(stakingtypes.Params{
					BondDenom: "BND",
				}, nil)
				stakingKeeper.EXPECT().GetValidator(gomock.Any(), gomock.Any()).Return(stakingtypes.Validator{
					Tokens: math.NewInt(0),
				}, nil)
				stakingKeeper.EXPECT().GetUnbondingDelegationsFromValidator(gomock.Any(), gomock.Any()).Return([]stakingtypes.UnbondingDelegation{}, nil)

				hooks := testutil.NewMockStakingHooks(ctrl)
				hooks.EXPECT().BeforeValidatorModified(gomock.Any(), gomock.Any()).Return(errors.New("hook failure"))
				stakingKeeper.EXPECT().Hooks().Return(hooks)

				stakingKeeper.EXPECT().RemoveValidatorTokens(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					stakingtypes.Validator{Status: stakingtypes.Bonded}, nil,
				)
				stakingKeeper.EXPECT().Unbond(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(math.ZeroInt(), nil)
			},
			bankMocks: func(_ sdk.Context, bankKeeper *testutil.MockBankKeeper) {
				bankKeeper.EXPECT().BurnCoins(gomock.Any(), stakingtypes.BondedPoolName, gomock.Any()).Return(nil)
			},
			expectedLog: "failed to call before validator modified hook",
		},
		{
			name:             "should pass - BeforeValidatorSlashed hook error is swallowed and logged",
			authority:        poaAuthority,
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
			stakingMocks: func(_ sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetParams(gomock.Any()).Return(stakingtypes.Params{
					BondDenom: "BND",
				}, nil)
				stakingKeeper.EXPECT().GetValidator(gomock.Any(), gomock.Any()).Return(stakingtypes.Validator{
					Tokens: sdk.DefaultPowerReduction,
				}, nil)
				stakingKeeper.EXPECT().GetUnbondingDelegationsFromValidator(gomock.Any(), gomock.Any()).Return([]stakingtypes.UnbondingDelegation{}, nil)

				hooks := testutil.NewMockStakingHooks(ctrl)
				hooks.EXPECT().BeforeValidatorModified(gomock.Any(), gomock.Any()).Return(nil)
				hooks.EXPECT().BeforeValidatorSlashed(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("hook failure"))
				stakingKeeper.EXPECT().Hooks().Return(hooks).Times(2)

				stakingKeeper.EXPECT().RemoveValidatorTokens(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					stakingtypes.Validator{Status: stakingtypes.Bonded}, nil,
				)
				stakingKeeper.EXPECT().Unbond(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(math.ZeroInt(), nil)
			},
			bankMocks: func(_ sdk.Context, bankKeeper *testutil.MockBankKeeper) {
				bankKeeper.EXPECT().BurnCoins(gomock.Any(), stakingtypes.BondedPoolName, gomock.Any()).Return(nil)
			},
			expectedLog: "failed to call before validator slashed hook",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			poaKeeper, ctx := setupPoaKeeper(t, tc.stakingMocks, tc.bankMocks)

			var logBuf bytes.Buffer
			if tc.expectedLog != "" {
				ctx = ctx.WithLogger(log.NewLogger(&logBuf, log.OutputJSONOption()))
			}

			msgServer := NewMsgServerImpl(*poaKeeper)

			msg := &types.MsgRemoveValidator{
				Authority:        tc.authority,
				ValidatorAddress: tc.validatorAddress,
			}

			_, err := msgServer.RemoveValidator(ctx, msg)
			if tc.expectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}
			if tc.expectedLog != "" {
				require.Contains(t, logBuf.String(), tc.expectedLog)
			}
		})
	}
}
