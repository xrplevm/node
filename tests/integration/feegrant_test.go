package integration

import (
	"cosmossdk.io/x/feegrant"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/evm/testutil/integration/base/factory"
	"github.com/stretchr/testify/require"
)

// x/feegrant was removed in v12, but its msgs still decode so that historical
// txs stay readable. A tx that fails to decode fails with ErrTxDecode, so the
// expected errors below also prove that the msgs decoded.
func (s *TestSuite) TestFeegrant_RemovedMsgs() {
	authority := sdktypes.AccAddress(address.Module("gov"))
	granter := s.keyring.GetKey(0)
	grantee := s.keyring.GetKey(1)

	allowance := &feegrant.BasicAllowance{}
	grant, err := feegrant.NewMsgGrantAllowance(allowance, granter.AccAddr, grantee.AccAddr)
	require.NoError(s.T(), err)
	revoke := feegrant.NewMsgRevokeAllowance(granter.AccAddr, grantee.AccAddr)
	prune := &feegrant.MsgPruneAllowances{Pruner: granter.AccAddr.String()}

	// Gov must sign every proposal msg, so gov is the granter or the pruner.
	govGrant, err := feegrant.NewMsgGrantAllowance(allowance, authority, grantee.AccAddr)
	require.NoError(s.T(), err)
	govRevoke := feegrant.NewMsgRevokeAllowance(authority, grantee.AccAddr)
	govPrune := &feegrant.MsgPruneAllowances{Pruner: authority.String()}
	proposal := func(msg sdktypes.Msg) sdktypes.Msg {
		p, err := govv1.NewMsgSubmitProposal(
			[]sdktypes.Msg{msg},
			sdktypes.NewCoins(sdktypes.NewCoin(s.Network().GetBaseDenom(), s.Network().GetMinDepositAmt())),
			granter.AccAddr.String(),
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
			name:          "feegrant - MsgGrantAllowance",
			msg:           grant,
			expectedError: sdkerrors.ErrUnknownRequest,
		},
		{
			name:          "feegrant - MsgRevokeAllowance",
			msg:           &revoke,
			expectedError: sdkerrors.ErrUnknownRequest,
		},
		{
			name:          "feegrant - MsgPruneAllowances",
			msg:           prune,
			expectedError: sdkerrors.ErrUnknownRequest,
		},
		{
			name:          "gov proposal - MsgGrantAllowance",
			msg:           proposal(govGrant),
			expectedError: govtypes.ErrUnroutableProposalMsg,
		},
		{
			name:          "gov proposal - MsgRevokeAllowance",
			msg:           proposal(&govRevoke),
			expectedError: govtypes.ErrUnroutableProposalMsg,
		},
		{
			name:          "gov proposal - MsgPruneAllowances",
			msg:           proposal(govPrune),
			expectedError: govtypes.ErrUnroutableProposalMsg,
		},
	}

	for _, tc := range tt {
		s.Run(tc.name, func() {
			gas := uint64(1_000_000)
			res, err := s.factory.CommitCosmosTx(granter.Priv, factory.CosmosTxArgs{
				Msgs: []sdktypes.Msg{tc.msg},
				Gas:  &gas,
			})
			require.NoError(s.T(), err)                                // The block did not fail.
			require.NotEqual(s.T(), uint32(0), res.Code)               // The tx failed.
			require.Contains(s.T(), res.Log, tc.expectedError.Error()) // The tx failed with the expected error.
		})
	}
}

// Without x/feegrant, the Cosmos ante handler must reject a tx that names a fee
// granter, so that nobody pays fees for another account.
func (s *TestSuite) TestFeegrant_FeeGranterRejected() {
	granter := s.keyring.GetKey(0)
	grantee := s.keyring.GetKey(1)
	denom := s.Network().GetBaseDenom()

	// balance reads the accounts that could pay the fee or receive the send.
	balance := func(addr sdktypes.AccAddress) sdktypes.Coin {
		res, err := s.Network().GetBankClient().Balance(
			s.Network().GetContext(),
			&banktypes.QueryBalanceRequest{Address: addr.String(), Denom: denom},
		)
		require.NoError(s.T(), err)
		return *res.Balance
	}
	granterBefore, granteeBefore := balance(granter.AccAddr), balance(grantee.AccAddr)

	// The grantee signs, and the granter is named as the account that pays the fee.
	gas := uint64(1_000_000)
	send := banktypes.NewMsgSend(grantee.AccAddr, granter.AccAddr, sdktypes.NewCoins(sdktypes.NewInt64Coin(denom, 1)))
	res, err := s.factory.CommitCosmosTx(grantee.Priv, factory.CosmosTxArgs{
		Msgs:       []sdktypes.Msg{send},
		Gas:        &gas,
		FeeGranter: granter.AccAddr,
	})
	require.NoError(s.T(), err)                                            // The block did not fail.
	require.Equal(s.T(), sdkerrors.ErrInvalidRequest.ABCICode(), res.Code) // The tx failed in the ante handler.
	require.Contains(s.T(), res.Log, "fee grants are not enabled")         // The tx failed because it names a fee granter.

	// Neither account paid a fee, and the send did not run.
	require.Equal(s.T(), granterBefore, balance(granter.AccAddr))
	require.Equal(s.T(), granteeBefore, balance(grantee.AccAddr))
}
