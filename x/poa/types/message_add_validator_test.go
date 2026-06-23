package types

import (
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestMsgAddValidator_ValidateBasic(t *testing.T) {
	validAddr := sdk.AccAddress("12345678901234567890").String()

	validPkAny, err := codectypes.NewAnyWithValue(ed25519.GenPrivKey().PubKey())
	require.NoError(t, err)

	wrongTypeAny, err := codectypes.NewAnyWithValue(&MsgRemoveValidator{})
	require.NoError(t, err)

	tt := []struct {
		name        string
		msg         *MsgAddValidator
		expectedErr string
	}{
		{
			name: "should pass - valid message",
			msg:  &MsgAddValidator{Authority: validAddr, ValidatorAddress: validAddr, Pubkey: validPkAny},
		},
		{
			name:        "should fail - invalid authority address",
			msg:         &MsgAddValidator{Authority: "invalid", ValidatorAddress: validAddr, Pubkey: validPkAny},
			expectedErr: "invalid authority address",
		},
		{
			name:        "should fail - invalid validator address",
			msg:         &MsgAddValidator{Authority: validAddr, ValidatorAddress: "invalid", Pubkey: validPkAny},
			expectedErr: "invalid validator address",
		},
		{
			name:        "should fail - nil pubkey",
			msg:         &MsgAddValidator{Authority: validAddr, ValidatorAddress: validAddr, Pubkey: nil},
			expectedErr: "validator pubkey is required",
		},
		{
			name:        "should fail - pubkey is not a cryptotypes.PubKey",
			msg:         &MsgAddValidator{Authority: validAddr, ValidatorAddress: validAddr, Pubkey: wrongTypeAny},
			expectedErr: "expecting cryptotypes.PubKey",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()

			if tc.expectedErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
