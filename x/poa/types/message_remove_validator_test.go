package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestMsgRemoveValidator_ValidateBasic(t *testing.T) {
	validAddr := sdk.AccAddress("12345678901234567890").String()

	tt := []struct {
		name        string
		authority   string
		validator   string
		expectedErr string
	}{
		{
			name:      "should pass - valid authority and validator address",
			authority: validAddr,
			validator: validAddr,
		},
		{
			name:        "should fail - invalid authority address",
			authority:   "invalid",
			validator:   validAddr,
			expectedErr: "invalid authority address",
		},
		{
			name:        "should fail - empty authority address",
			authority:   "",
			validator:   validAddr,
			expectedErr: "invalid authority address",
		},
		{
			name:        "should fail - invalid validator address",
			authority:   validAddr,
			validator:   "invalid",
			expectedErr: "invalid validator address",
		},
		{
			name:        "should fail - empty validator address",
			authority:   validAddr,
			validator:   "",
			expectedErr: "invalid validator address",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			msg := NewMsgRemoveValidator(tc.authority, tc.validator)

			err := msg.ValidateBasic()

			if tc.expectedErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
