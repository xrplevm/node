package keeper

import (
	"errors"
	"testing"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"
	"github.com/xrplevm/node/v5/x/poa/types"
)

func TestMsgServer_RemoveValidator(t *testing.T) {
	poaKeeper, ctx := poaKeeperTestSetup(t)

	msgServer := NewMsgServerImpl(*poaKeeper)

	tt := []struct{
		name string
		authority string
		validatorAddress string
		expectedErr error
	}{
		{
			name: "should fail - invalid authority address",
			authority: "invalidauthority",
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
			expectedErr: govtypes.ErrInvalidSigner,
		},
		{
			name: "should fail - invalid validator address",
			authority: poaKeeper.GetAuthority(),
			validatorAddress: "invalidvalidatoraddress",
			expectedErr: errors.New("decoding bech32 failed"),
		},
		{
			name:  "should pass",
			authority: poaKeeper.GetAuthority(),
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
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
		})
	}

}
