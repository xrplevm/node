package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xrplevm/node/v6/x/poa/types"
)

func TestMsgServer_RemoveValidator(t *testing.T) {
	poaKeeper, ctx := poaKeeperTestSetup(t)

	msgServer := NewMsgServerImpl(*poaKeeper)

	msg := &types.MsgRemoveValidator{
		Authority:        poaKeeper.GetAuthority(),
		ValidatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
	}

	_, err := msgServer.RemoveValidator(ctx, msg)
	require.NoError(t, err)
}
