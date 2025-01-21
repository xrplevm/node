package keeper

import (
	"errors"
	"testing"

	types1 "github.com/cosmos/cosmos-sdk/codec/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
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

	tt := []struct {
		name             string
		authority        string
		validatorAddress string
		expectedErr      error
	}{
		{
			name:             "should fail - invalid authority address",
			authority:        "invalidauthority",
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
			expectedErr:      govtypes.ErrInvalidSigner,
		},
		{
			name:             "should fail - invalid validator address",
			authority:        poaKeeper.GetAuthority(),
			validatorAddress: "invalidvalidatoraddress",
			expectedErr:      errors.New("decoding bech32 failed"),
		},
		{
			name:             "should pass",
			authority:        poaKeeper.GetAuthority(),
			validatorAddress: "ethm1a0pd5cyew47pvgf7rd7axxy3humv9ev0nnkprp",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			msg := &types.MsgAddValidator{
				Authority:        tc.authority,
				ValidatorAddress: tc.validatorAddress,
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
			if tc.expectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
