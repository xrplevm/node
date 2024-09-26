package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestPoA_Hooks(t *testing.T) {
	keeper, ctx, _ := setupPoAKeeper(t)

	hooks := keeper.Hooks()

	err := hooks.BeforeDelegationCreated(ctx, sdk.AccAddress("ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp"), sdk.ValAddress("ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp"))
	require.NoError(t, err)
	err = hooks.AfterValidatorBonded(ctx, sdk.ConsAddress(""), sdk.ValAddress(""))
	require.NoError(t, err)
	err = hooks.AfterValidatorRemoved(ctx, sdk.ConsAddress(""), sdk.ValAddress(""))
	require.NoError(t, err)
	err = hooks.AfterValidatorCreated(ctx, sdk.ValAddress(""))
	require.NoError(t, err)
	err = hooks.AfterValidatorBeginUnbonding(ctx, sdk.ConsAddress(""), sdk.ValAddress(""))
	require.NoError(t, err)
	err = hooks.BeforeValidatorModified(ctx, sdk.ValAddress(""))
	require.NoError(t, err)
	err = hooks.BeforeDelegationSharesModified(ctx, sdk.AccAddress(""), sdk.ValAddress(""))
	require.NoError(t, err)
	err = hooks.BeforeDelegationRemoved(ctx, sdk.AccAddress(""), sdk.ValAddress(""))
	require.NoError(t, err)
	err = hooks.AfterDelegationModified(ctx, sdk.AccAddress(""), sdk.ValAddress(""))
	require.NoError(t, err)
	err = hooks.AfterUnbondingInitiated(ctx, 0)
	require.NoError(t, err)
	err = hooks.BeforeValidatorSlashed(ctx, sdk.ValAddress(""), sdk.NewDec(0))
	require.NoError(t, err)
}
