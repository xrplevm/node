package keeper

import (
	"errors"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	"github.com/xrplevm/node/v10/x/poa/testutil"
	"github.com/xrplevm/node/v10/x/poa/types"
	"go.uber.org/mock/gomock"
)

func TestMsgServer_RemoveValidatorSelf(t *testing.T) {
	ctrl := gomock.NewController(t)

	const validatorAddress = "ethm1wunfhl05vc8r8xxnnp8gt62wa54r6y52pg03zq"
	bondedValidator := stakingtypes.Validator{Status: stakingtypes.Bonded, Tokens: math.NewInt(0)}

	tt := []struct {
		name             string
		validatorAddress string
		stakingMocks     func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper)
		bankMocks        func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper)
		expectedErr      error
	}{
		{
			name:             "should fail - invalid validator address",
			validatorAddress: "invalidvalidatoraddress",
			stakingMocks:     func(_ sdk.Context, _ *testutil.MockStakingKeeper) {},
			bankMocks:        func(_ sdk.Context, _ *testutil.MockBankKeeper) {},
			expectedErr:      errors.New("decoding bech32 failed"),
		},
		{
			name:             "should fail - address is not a validator",
			validatorAddress: validatorAddress,
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(stakingtypes.Validator{}, errors.New("validator not found"))
			},
			bankMocks:   func(_ sdk.Context, _ *testutil.MockBankKeeper) {},
			expectedErr: types.ErrAddressIsNotAValidator,
		},
		{
			name:             "should fail - last bonded validator cannot self-remove",
			validatorAddress: validatorAddress,
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(bondedValidator, nil)
				stakingKeeper.EXPECT().GetAllValidators(ctx).Return([]stakingtypes.Validator{bondedValidator}, nil)
			},
			bankMocks:   func(_ sdk.Context, _ *testutil.MockBankKeeper) {},
			expectedErr: types.ErrCannotRemoveLastValidator,
		},
		{
			name:             "should pass - reuses the RemoveValidator logic",
			validatorAddress: validatorAddress,
			stakingMocks: func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
				stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).Return(bondedValidator, nil).Times(2)
				stakingKeeper.EXPECT().GetAllValidators(ctx).Return([]stakingtypes.Validator{bondedValidator, bondedValidator}, nil)
				stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{BondDenom: "BND"}, nil)
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
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			poaKeeper, ctx := setupPoaKeeper(t, tc.stakingMocks, tc.bankMocks)

			msgServer := NewMsgServerImpl(*poaKeeper)

			msg := &types.MsgRemoveValidatorSelf{
				ValidatorAddress: tc.validatorAddress,
			}

			_, err := msgServer.RemoveValidatorSelf(ctx, msg)
			if tc.expectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
