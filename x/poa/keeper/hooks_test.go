package keeper

import (
	"testing"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/xrplevm/node/v5/x/poa/testutil"
)

func TestHooks_BeforeDelegationCreated(t *testing.T) {
	keeper, ctx := setupPoaKeeper(
		t,
		func(_ sdk.Context, _ *testutil.MockStakingKeeper) {},
		func(_ sdk.Context, _ *testutil.MockBankKeeper) {},
		func(_ sdk.Context, _ *testutil.MockSlashingKeeper) {},
	)

	hooks := keeper.Hooks()

	// Test delegating to self (should succeed)
	addr := sdk.AccAddress("ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp")
	valAddr := sdk.ValAddress(addr)
	err := hooks.BeforeDelegationCreated(ctx, addr, valAddr)
	require.NoError(t, err)

	// Test delegating to other (should fail)
	otherValAddr := sdk.ValAddress("ethm1other47pvgf7rd7axxy3humv9ev0nnkprp")
	err = hooks.BeforeDelegationCreated(ctx, addr, otherValAddr)
	require.Error(t, err)
	require.Contains(t, err.Error(), "delegation to other accounts is not allowed")
}

func TestHooks_AfterValidatorBonded(t *testing.T) {
	keeper, ctx := setupPoaKeeper(
		t,
		func(_ sdk.Context, _ *testutil.MockStakingKeeper) {},
		func(_ sdk.Context, _ *testutil.MockBankKeeper) {},
		func(_ sdk.Context, _ *testutil.MockSlashingKeeper) {},
	)

	hooks := keeper.Hooks()

	err := hooks.AfterValidatorBonded(ctx, sdk.ConsAddress("test"), sdk.ValAddress("test"))
	require.NoError(t, err)
}

func TestHooks_AfterValidatorRemoved(t *testing.T) {
	keeper, ctx := setupPoaKeeper(
		t,
		func(_ sdk.Context, _ *testutil.MockStakingKeeper) {},
		func(_ sdk.Context, _ *testutil.MockBankKeeper) {},
		func(_ sdk.Context, _ *testutil.MockSlashingKeeper) {},
	)

	hooks := keeper.Hooks()

	err := hooks.AfterValidatorRemoved(ctx, sdk.ConsAddress("test"), sdk.ValAddress("test"))
	require.NoError(t, err)
}

func TestHooks_AfterValidatorCreated(t *testing.T) {
	keeper, ctx := setupPoaKeeper(
		t,
		func(_ sdk.Context, _ *testutil.MockStakingKeeper) {},
		func(_ sdk.Context, _ *testutil.MockBankKeeper) {},
		func(_ sdk.Context, _ *testutil.MockSlashingKeeper) {},
	)

	hooks := keeper.Hooks()

	err := hooks.AfterValidatorCreated(ctx, sdk.ValAddress("test"))
	require.NoError(t, err)
}

func TestHooks_AfterValidatorBeginUnbonding(t *testing.T) {
	keeper, ctx := setupPoaKeeper(
		t,
		func(_ sdk.Context, _ *testutil.MockStakingKeeper) {},
		func(_ sdk.Context, _ *testutil.MockBankKeeper) {},
		func(_ sdk.Context, _ *testutil.MockSlashingKeeper) {},
	)

	hooks := keeper.Hooks()

	err := hooks.AfterValidatorBeginUnbonding(ctx, sdk.ConsAddress("test"), sdk.ValAddress("test"))
	require.NoError(t, err)
}

func TestHooks_BeforeValidatorModified(t *testing.T) {
	keeper, ctx := setupPoaKeeper(
		t,
		func(_ sdk.Context, _ *testutil.MockStakingKeeper) {},
		func(_ sdk.Context, _ *testutil.MockBankKeeper) {},
		func(_ sdk.Context, _ *testutil.MockSlashingKeeper) {},
	)

	hooks := keeper.Hooks()

	err := hooks.BeforeValidatorModified(ctx, sdk.ValAddress("test"))
	require.NoError(t, err)
}

func TestHooks_BeforeDelegationSharesModified(t *testing.T) {
	keeper, ctx := setupPoaKeeper(
		t,
		func(_ sdk.Context, _ *testutil.MockStakingKeeper) {},
		func(_ sdk.Context, _ *testutil.MockBankKeeper) {},
		func(_ sdk.Context, _ *testutil.MockSlashingKeeper) {},
	)

	hooks := keeper.Hooks()

	err := hooks.BeforeDelegationSharesModified(ctx, sdk.AccAddress("test"), sdk.ValAddress("test"))
	require.NoError(t, err)
}

func TestHooks_BeforeDelegationRemoved(t *testing.T) {
	keeper, ctx := setupPoaKeeper(
		t,
		func(_ sdk.Context, _ *testutil.MockStakingKeeper) {},
		func(_ sdk.Context, _ *testutil.MockBankKeeper) {},
		func(_ sdk.Context, _ *testutil.MockSlashingKeeper) {},
	)

	hooks := keeper.Hooks()

	err := hooks.BeforeDelegationRemoved(ctx, sdk.AccAddress("test"), sdk.ValAddress("test"))
	require.NoError(t, err)
}

func TestHooks_AfterDelegationModified(t *testing.T) {
	keeper, ctx := setupPoaKeeper(
		t,
		func(_ sdk.Context, _ *testutil.MockStakingKeeper) {},
		func(_ sdk.Context, _ *testutil.MockBankKeeper) {},
		func(_ sdk.Context, _ *testutil.MockSlashingKeeper) {},
	)

	hooks := keeper.Hooks()

	err := hooks.AfterDelegationModified(ctx, sdk.AccAddress("test"), sdk.ValAddress("test"))
	require.NoError(t, err)
}

func TestHooks_AfterUnbondingInitiated(t *testing.T) {
	keeper, ctx := setupPoaKeeper(
		t,
		func(_ sdk.Context, _ *testutil.MockStakingKeeper) {},
		func(_ sdk.Context, _ *testutil.MockBankKeeper) {},
		func(_ sdk.Context, _ *testutil.MockSlashingKeeper) {},
	)

	hooks := keeper.Hooks()

	err := hooks.AfterUnbondingInitiated(ctx, 123)
	require.NoError(t, err)
}

func TestHooks_BeforeValidatorSlashed(t *testing.T) {
	keeper, ctx := setupPoaKeeper(
		t,
		func(_ sdk.Context, _ *testutil.MockStakingKeeper) {},
		func(_ sdk.Context, _ *testutil.MockBankKeeper) {},
		func(_ sdk.Context, _ *testutil.MockSlashingKeeper) {},
	)

	hooks := keeper.Hooks()

	err := hooks.BeforeValidatorSlashed(ctx, sdk.ValAddress("test"), math.LegacyNewDec(100))
	require.NoError(t, err)
}
