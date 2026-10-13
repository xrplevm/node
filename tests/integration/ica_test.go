package integration

import (
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/evm/testutil/integration/base/factory"
	icacontrollertypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	"github.com/stretchr/testify/require"
)

// ICA host was removed in v12, but every ICS-27 msg still decodes so that
// historical txs stay readable. A tx that fails to decode fails with
// ErrTxDecode, so the expected errors below also prove that the msgs decoded.
func (s *TestSuite) TestICA_RemovedMsgs() {
	authority := sdktypes.AccAddress(address.Module("gov")).String()
	sender := s.keyring.GetKey(0)

	hostParams := icahosttypes.NewParams(true, []string{"*"})
	controllerParams := icacontrollertypes.NewParams(true)
	query := []icahosttypes.QueryRequest{{Path: "/cosmos.bank.v1beta1.Query/AllBalances", Data: []byte{0x0a, 0x01, 0x01}}}
	packetData := icatypes.InterchainAccountPacketData{Type: icatypes.EXECUTE_TX, Data: []byte{0x01}}
	timeout := uint64(600_000_000_000)

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
			name:          "ica host - MsgUpdateParams",
			msg:           icahosttypes.NewMsgUpdateParams(sender.AccAddr.String(), hostParams),
			expectedError: sdkerrors.ErrUnknownRequest,
		},
		{
			name:          "ica host - MsgModuleQuerySafe",
			msg:           icahosttypes.NewMsgModuleQuerySafe(sender.AccAddr.String(), query),
			expectedError: sdkerrors.ErrUnknownRequest,
		},
		{
			name:          "ica controller - MsgRegisterInterchainAccount",
			msg:           icacontrollertypes.NewMsgRegisterInterchainAccount("connection-0", sender.AccAddr.String(), "", channeltypes.ORDERED),
			expectedError: sdkerrors.ErrUnknownRequest,
		},
		{
			name:          "ica controller - MsgSendTx",
			msg:           icacontrollertypes.NewMsgSendTx(sender.AccAddr.String(), "connection-0", timeout, packetData),
			expectedError: sdkerrors.ErrUnknownRequest,
		},
		{
			name:          "ica controller - MsgUpdateParams",
			msg:           icacontrollertypes.NewMsgUpdateParams(sender.AccAddr.String(), controllerParams),
			expectedError: sdkerrors.ErrUnknownRequest,
		},
		{
			name:          "gov proposal - ica host MsgUpdateParams",
			msg:           proposal(icahosttypes.NewMsgUpdateParams(authority, hostParams)),
			expectedError: govtypes.ErrUnroutableProposalMsg,
		},
		{
			name:          "gov proposal - ica host MsgModuleQuerySafe",
			msg:           proposal(icahosttypes.NewMsgModuleQuerySafe(authority, query)),
			expectedError: govtypes.ErrUnroutableProposalMsg,
		},
		{
			name:          "gov proposal - ica controller MsgRegisterInterchainAccount",
			msg:           proposal(icacontrollertypes.NewMsgRegisterInterchainAccount("connection-0", authority, "", channeltypes.ORDERED)),
			expectedError: govtypes.ErrUnroutableProposalMsg,
		},
		{
			name:          "gov proposal - ica controller MsgSendTx",
			msg:           proposal(icacontrollertypes.NewMsgSendTx(authority, "connection-0", timeout, packetData)),
			expectedError: govtypes.ErrUnroutableProposalMsg,
		},
		{
			name:          "gov proposal - ica controller MsgUpdateParams",
			msg:           proposal(icacontrollertypes.NewMsgUpdateParams(authority, controllerParams)),
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
