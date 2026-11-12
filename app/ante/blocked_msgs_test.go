package ante

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	"github.com/xrplevm/node/v12/x/poa/testutil"
	"go.uber.org/mock/gomock"
)

func TestBlockedMsgsDecorator_AnteHandle(t *testing.T) {
	proposalWithBlockedMsg, err := govv1.NewMsgSubmitProposal(
		[]sdk.Msg{&distrtypes.MsgFundCommunityPool{}}, nil, "", "", "test", "test", false,
	)
	require.NoError(t, err)

	tt := []struct {
		name          string
		msgs          []sdk.Msg
		expectedError error
	}{
		{
			name:          "should return error - rejects a blocked msg",
			msgs:          []sdk.Msg{&distrtypes.MsgFundCommunityPool{}},
			expectedError: errortypes.ErrUnauthorized,
		},
		{
			name:          "should return error - rejects a blocked msg after an allowed msg",
			msgs:          []sdk.Msg{&stakingtypes.MsgEditValidator{}, &govv1.MsgVoteWeighted{}},
			expectedError: errortypes.ErrUnauthorized,
		},
		{
			name: "pass - allows a msg that is not blocked",
			msgs: []sdk.Msg{&stakingtypes.MsgEditValidator{}},
		},
		{
			name: "pass - allows a gov proposal that carries a blocked msg",
			msgs: []sdk.Msg{proposalWithBlockedMsg},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			txMock := testutil.NewMockTx(ctrl)
			txMock.EXPECT().GetMsgs().Return(tc.msgs).AnyTimes()

			mockNext := func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
				return ctx, nil
			}

			_, err := blockedMsgsDecorator{}.AnteHandle(sdk.Context{}, txMock, false, mockNext)
			if tc.expectedError != nil {
				require.ErrorIs(t, err, tc.expectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
