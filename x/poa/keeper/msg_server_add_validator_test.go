package keeper

import (
	"testing"

	types1 "github.com/cosmos/cosmos-sdk/codec/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/xrplevm/node/v6/x/poa/testutil"
	"github.com/xrplevm/node/v6/x/poa/types"
)

func TestMsgServer_AddValidator(t *testing.T) {
	poaKeeper, ctx := poaKeeperTestSetup(t)

	ctrl := gomock.NewController(t)
	pubKey := testutil.NewMockPubKey(ctrl)
	msgPubKey, _ := types1.NewAnyWithValue(pubKey)
	msgServer := NewMsgServerImpl(*poaKeeper)

	msg := &types.MsgAddValidator{
		Authority:        poaKeeper.GetAuthority(),
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

	_, err := msgServer.AddValidator(ctx, msg)
	require.NoError(t, err)
}
