package keeper

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	"github.com/xrplevm/node/v3/x/poa/testutil"
)

func TestStakingPowerInvariant_Valid(t *testing.T) {
	poaKeeper, ctx := setupPoaKeeper(
		t,
		func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
			stakingKeeper.EXPECT().GetAllValidators(ctx).Return([]stakingtypes.Validator{
				{
					Tokens: sdk.DefaultPowerReduction,
				},
				{
					Tokens: math.ZeroInt(),
				},
			}, nil)
		},
		func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {},
		func(ctx sdk.Context, slashingKeeper *testutil.MockSlashingKeeper) {},
	)

	invariant := StakingPowerInvariant(*poaKeeper)
	msg, broken := invariant(ctx)
	require.False(t, broken, msg)
}

func TestStakingPowerInvariant_Invalid(t *testing.T) {
	poaKeeper, ctx := setupPoaKeeper(
		t,
		func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {
			stakingKeeper.EXPECT().GetAllValidators(ctx).Return([]stakingtypes.Validator{
				{
					Tokens: sdk.DefaultPowerReduction,
				},
				{
					Tokens: sdk.DefaultPowerReduction.Add(math.OneInt()),
				},
			}, nil)
		},
		func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {},
		func(ctx sdk.Context, slashingKeeper *testutil.MockSlashingKeeper) {},
	)

	invariant := StakingPowerInvariant(*poaKeeper)
	msg, broken := invariant(ctx)
	require.True(t, broken, msg)
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
		func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {},
		func(ctx sdk.Context, slashingKeeper *testutil.MockSlashingKeeper) {},
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
		func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {},
		func(ctx sdk.Context, slashingKeeper *testutil.MockSlashingKeeper) {},
	)

	invariant := SelfDelegationInvariant(*poaKeeper)
	msg, broken := invariant(ctx)
	require.True(t, broken, msg)
}

func TestCheckKeeperDependenciesParamsInvariant_Valid(t *testing.T) {
	poaKeeper, ctx := setupPoaKeeper(
		t,
		func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {},
		func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {},
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
		func(ctx sdk.Context, stakingKeeper *testutil.MockStakingKeeper) {},
		func(ctx sdk.Context, bankKeeper *testutil.MockBankKeeper) {},
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
