package integration

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/cosmos/evm/testutil/integration/base/factory"
	"github.com/stretchr/testify/require"
)

// x/params was removed in v12, but ParameterChangeProposal still decodes so
// that historical txs and gov proposals stay readable. Gov has no handler for
// it, so it must reject it at submission. A tx that fails to decode fails with
// ErrTxDecode, so the expected errors below also prove that the proposals
// decoded.
func (s *TestSuite) TestParameterChangeProposal_NotRouted() {
	authority := sdktypes.AccAddress(address.Module("gov")).String()
	sender := s.keyring.GetKey(0)

	// paramChange builds the msg that the removed param-change CLI built: one
	// subspace change, wrapped as legacy proposal content that gov executes.
	paramChange := func(subspace, key, value string) sdktypes.Msg {
		content, err := codectypes.NewAnyWithValue(paramproposal.NewParameterChangeProposal("test", "test",
			[]paramproposal.ParamChange{paramproposal.NewParamChange(subspace, key, value)}))
		require.NoError(s.T(), err)
		return govv1.NewMsgExecLegacyContent(content, authority)
	}
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

	// Before v12, gov accepted these targets: their subspaces had registered keys.
	tt := []struct {
		name          string
		msg           sdktypes.Msg
		expectedError error
	}{
		{
			name:          "gov proposal - transfer SendEnabled",
			msg:           proposal(paramChange("transfer", "SendEnabled", "false")),
			expectedError: govtypes.ErrNoProposalHandlerExists,
		},
		{
			name:          "gov proposal - gov votingparams",
			msg:           proposal(paramChange("gov", "votingparams", `{"voting_period":"60000000000"}`)),
			expectedError: govtypes.ErrNoProposalHandlerExists,
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
