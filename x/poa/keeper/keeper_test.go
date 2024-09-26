package keeper

import (
	"testing"

	types1 "github.com/cosmos/cosmos-sdk/codec/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	"github.com/xrplevm/node/v3/x/poa/types"
)

// Define here Keeper methods to be unit tested
func TestPoAKeeper_ExecuteAddValidator(t *testing.T) {
	keeper, ctx, pubKey := setupPoAKeeper(t)

	msgPubKey, _ := types1.NewAnyWithValue(pubKey)

	msg := &types.MsgAddValidator{
		Authority:        keeper.GetAuthority(),
		ValidatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
		Description: stakingtypes.Description{
			Moniker:         "test",
			Identity:        "test",
			Website:         "test",
			SecurityContact: "test",
			Details:         "test",
		},
		Pubkey: msgPubKey,
	}

	err := keeper.ExecuteAddValidator(ctx, msg)
	require.NoError(t, err)
}

func TestPoAKeeper_ExecuteRemoveValidator(t *testing.T) {
	keeper, ctx, _ := setupPoAKeeper(t)

	err := keeper.ExecuteRemoveValidator(ctx, "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp")
	require.NoError(t, err)
}
