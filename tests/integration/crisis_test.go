package integration

import (
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/evm/testutil/integration/base/factory"
	"github.com/stretchr/testify/require"
)

// x/crisis was removed in v12, but its msgs still decode so that historical
// txs stay readable. A tx that fails to decode fails with ErrTxDecode, so the
// expected errors below also prove that the msgs decoded.
func (s *TestSuite) TestCrisis_RemovedMsgs() {
	authority := sdktypes.AccAddress(address.Module("gov")).String()
	sender := s.keyring.GetKey(0)

	constantFee := sdktypes.NewInt64Coin(s.Network().GetBaseDenom(), 1)
	proposal := func(msg sdktypes.Msg) sdktypes.Msg {
		p, err := govv1.NewMsgSubmitProposal(
			[]sdktypes.Msg{msg},
			sdktypes.NewCoins(sdktypes.NewCoin(s.Network().GetBaseDenom(), s.Network().GetMinDepositAmt())),
			sender.AccAddr.String(),
			"",
			"test",
			"test",
			false,
		)
		require.NoError(s.T(), err)
		return p
	}

	tt := []struct {
		name          string
		msg           sdktypes.Msg
		expectedError error
	}{
		{
			name:          "crisis - MsgVerifyInvariant",
			msg:           crisistypes.NewMsgVerifyInvariant(sender.AccAddr, "bank", "total-supply"),
			expectedError: sdkerrors.ErrUnknownRequest,
		},
		{
			name:          "crisis - MsgUpdateParams",
			msg:           &crisistypes.MsgUpdateParams{Authority: sender.AccAddr.String(), ConstantFee: constantFee},
			expectedError: sdkerrors.ErrUnknownRequest,
		},
		{
			name:          "gov proposal - MsgUpdateParams",
			msg:           proposal(&crisistypes.MsgUpdateParams{Authority: authority, ConstantFee: constantFee}),
			expectedError: govtypes.ErrUnroutableProposalMsg,
		},
	}

	for _, tc := range tt {
		s.Run(tc.name, func() {
			gas := uint64(1_000_000)
			res, err := s.factory.CommitCosmosTx(sender.Priv, factory.CosmosTxArgs{
				Msgs: []sdktypes.Msg{tc.msg},
				Gas:  &gas,
			})
			require.NoError(s.T(), err)                                // The block did not fail.
			require.NotEqual(s.T(), uint32(0), res.Code)               // The tx failed.
			require.Contains(s.T(), res.Log, tc.expectedError.Error()) // The tx failed with the expected error.
		})
	}
}
