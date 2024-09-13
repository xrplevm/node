package keeper

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	"github.com/xrplevm/node/v3/x/poa/testutil"
)

const (
	mockDenom = "test"
)

var (
	mockCoin = sdk.Coin{
		Denom: mockDenom,
	}
)

func TestCheckValidatorStakingPower_CoinDenomDoesNotMatchBondDenom(t *testing.T) {
	var (
		msg string
		broken bool
	)

	poaKeeper, ctx := setupPoaKeeper(
		t,
		func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
			stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
				BondDenom: "mockDenom",
			}).AnyTimes()
		},
		func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {},
	)

	checkValidatorStakingPower(ctx, *poaKeeper, &msg, &broken)(sdk.AccAddress(""), sdk.NewCoin("unmatched", math.NewInt(100)))

	require.False(t, broken)
}

func TestCheckValidatorStakingPower_ValidatorNotFound(t *testing.T) {
	poaKeeper, ctx := setupPoaKeeper(
		t,
		func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
			stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
				BondDenom: mockDenom,
			}).AnyTimes()
			stakingKeeper.EXPECT().GetValidator(ctx, sdk.ValAddress("")).Return(stakingtypes.Validator{}, false)
		},
		func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {},
	)

	var (
		msg string
		broken bool
	)

	checkValidatorStakingPower(ctx, *poaKeeper, &msg, &broken)(sdk.AccAddress(""), mockCoin)

	require.False(t, broken)
}

func TestCheckValidatorStakingPower_ValidatorMatchesStakingPower(t *testing.T) {
	var (
		msg string
		broken bool
	)

	poaKeeper, ctx := setupPoaKeeper(
		t,
		func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
			stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
				BondDenom: mockDenom,
			}).AnyTimes()
			stakingKeeper.EXPECT().GetValidator(ctx, sdk.ValAddress("")).Return(stakingtypes.Validator{
				Tokens: sdk.DefaultPowerReduction,
			}, true).AnyTimes()
		},
		func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {},
	)

	checkValidatorStakingPower(ctx, *poaKeeper, &msg, &broken)(sdk.AccAddress(""), mockCoin)

	require.False(t, broken)
}

func TestCheckValidatorStakingPower_ValidatorDoesNotMatchStakingPower(t *testing.T) {
	var (
		msg string
		broken bool
	)

	poaKeeper, ctx := setupPoaKeeper(
		t,
		func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
			stakingKeeper.EXPECT().GetParams(ctx).Return(stakingtypes.Params{
				BondDenom: mockDenom,
			}).AnyTimes()
			stakingKeeper.EXPECT().GetValidator(ctx, sdk.ValAddress("")).Return(stakingtypes.Validator{
				Tokens: math.NewInt(100),
			}, true).AnyTimes()
		},
		func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {},
	)

	checkValidatorStakingPower(ctx, *poaKeeper, &msg, &broken)(sdk.AccAddress(""), mockCoin)

	require.True(t, broken)
}


