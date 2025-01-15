package keeper

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	"github.com/xrplevm/node/v5/x/poa/testutil"
)

func TestStakingPowerInvariant_Valid(t *testing.T) {
	tt := []struct {
		name       string
		broken     bool
		validators func() ([]stakingtypes.Validator, error)
	}{
		{
			name:   "should pass - all validators have the same staking power",
			broken: false,
			validators: func() ([]stakingtypes.Validator, error) {
				return []stakingtypes.Validator{
					{
						Tokens: sdk.DefaultPowerReduction,
					},
				}, nil
			},
		},
		{
			name:   "should fail - one validator has excessive staking power",
			broken: true,
			validators: func() ([]stakingtypes.Validator, error) {
				return []stakingtypes.Validator{
					{
						Tokens: sdk.DefaultPowerReduction.Add(math.OneInt()),
					},
				}, nil
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			poaKeeper, ctx := setupPoaKeeper(
				t,
				func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
					stakingKeeper.EXPECT().GetAllValidators(ctx).Return(tc.validators())
				},
				func(sdk.Context, *testutil.MockBankKeeper) {},
				func(sdk.Context, *testutil.MockSlashingKeeper) {},
			)

			invariant := StakingPowerInvariant(*poaKeeper)
			_, broken := invariant(ctx)
			require.Equal(t, broken, tc.broken)
		})
	}
}

func TestSelfDelegationInvariant_Valid(t *testing.T) {
	poaKeeper, ctx := setupPoaKeeper(
		t,
		func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
			stakingKeeper.EXPECT().GetAllDelegations(ctx).Return([]stakingtypes.Delegation{
				{
					DelegatorAddress: "ethm13ued6aqj3w7jvks4l270dunhue0a9y7tspnpn5",
					ValidatorAddress: "ethmvaloper13ued6aqj3w7jvks4l270dunhue0a9y7tl3edtf",
				},
				{
					DelegatorAddress: "ethm13ued6aqj3w7jvks4l270dunhue0a9y7tspnpn5",
					ValidatorAddress: "ethmvaloper13ued6aqj3w7jvks4l270dunhue0a9y7tl3edtf",
				},
			}, nil)
		},
		func(_ sdk.Context, _ *testutil.MockBankKeeper) {},
		func(_ sdk.Context, _ *testutil.MockSlashingKeeper) {},
	)

	invariant := SelfDelegationInvariant(*poaKeeper)
	msg, broken := invariant(ctx)
	require.False(t, broken, msg)
}

func TestSelfDelegationInvariant_Invalid(t *testing.T) {
	poaKeeper, ctx := setupPoaKeeper(
		t,
		func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
			stakingKeeper.EXPECT().GetAllDelegations(ctx).Return([]stakingtypes.Delegation{
				{
					DelegatorAddress: "ethm1wunfhl05vc8r8xxnnp8gt62wa54r6y52pg03zq",
					ValidatorAddress: "ethmvaloper13ued6aqj3w7jvks4l270dunhue0a9y7tl3edtf",
				},
			}, nil)
		},
		func(_ sdk.Context, _ *testutil.MockBankKeeper) {},
		func(_ sdk.Context, _ *testutil.MockSlashingKeeper) {},
	)

	invariant := SelfDelegationInvariant(*poaKeeper)
	msg, broken := invariant(ctx)
	require.True(t, broken, msg)
}

func TestCheckKeeperDependenciesParamsInvariant_Valid(t *testing.T) {
	poaKeeper, ctx := setupPoaKeeper(
		t,
		func(_ sdk.Context, _ *testutil.MockStakingKeeper) {},
		func(_ sdk.Context, _ *testutil.MockBankKeeper) {},
		func(ctx sdk.Context, slashingKeeper *testutil.MockSlashingKeeper) {
			slashingKeeper.EXPECT().GetParams(ctx).Return(slashingtypes.Params{
				SlashFractionDoubleSign: math.LegacyZeroDec(),
				SlashFractionDowntime:   math.LegacyZeroDec(),
			}, nil)
		},
	)

	invariant := CheckKeeperDependenciesParamsInvariant(*poaKeeper)
	msg, broken := invariant(ctx)
	require.False(t, broken, msg)
}

func TestCheckKeeperDependenciesParamsInvariant_Invalid(t *testing.T) {
	poaKeeper, ctx := setupPoaKeeper(
		t,
		func(_ sdk.Context, _ *testutil.MockStakingKeeper) {},
		func(_ sdk.Context, _ *testutil.MockBankKeeper) {},
		func(ctx sdk.Context, slashingKeeper *testutil.MockSlashingKeeper) {
			slashingKeeper.EXPECT().GetParams(ctx).Return(slashingtypes.Params{
				SlashFractionDoubleSign: math.LegacyNewDecWithPrec(5, 2), // 0.05
				SlashFractionDowntime:   math.LegacyNewDecWithPrec(6, 1), // 0.6 (invalid, should be less than MinSignedPerWindow)
			}, nil)
		},
	)

	invariant := CheckKeeperDependenciesParamsInvariant(*poaKeeper)
	msg, broken := invariant(ctx)
	require.True(t, broken, msg)
}
