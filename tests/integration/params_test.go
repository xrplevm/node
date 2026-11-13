package integration

import (
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/cosmos/evm/testutil/integration/base/factory"
	"github.com/stretchr/testify/require"
)

// x/params was removed in v12, but ParameterChangeProposal still decodes so
// that historical txs and gov proposals stay readable. The legacy gov router
// has no params route, so gov must reject this content at submission. A tx that
// fails to decode fails with ErrTxDecode, so the expected errors below also
// prove that the proposals decoded.
func (s *TestSuite) TestParameterChangeProposal_NotRouted() {
	authority := sdktypes.AccAddress(address.Module("gov")).String()
	sender := s.keyring.GetKey(0)

	legacyContent := func(content govv1beta1.Content) sdktypes.Msg {
		msg, err := govv1.NewLegacyContent(content, authority)
		require.NoError(s.T(), err)
		return msg
	}
	// paramChange builds the msg that the removed param-change CLI built: one
	// subspace change, wrapped as legacy proposal content that gov executes.
	paramChange := func(subspace, key, value string) sdktypes.Msg {
		return legacyContent(paramproposal.NewParameterChangeProposal("test", "test",
			[]paramproposal.ParamChange{paramproposal.NewParamChange(subspace, key, value)}))
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

	// Before v12, gov accepted this target: its subspace had registered keys.
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
		// Control: the gov route of the legacy router still works, so the
		// rejection above comes from the missing params route.
		{
			name: "gov proposal - legacy text",
			msg:  proposal(legacyContent(govv1beta1.NewTextProposal("test", "test"))),
		},
	}

	for _, tc := range tt {
		s.Run(tc.name, func() {
			gas := uint64(1_000_000)
			res, err := s.factory.CommitCosmosTx(sender.Priv, factory.CosmosTxArgs{
				Msgs: []sdktypes.Msg{tc.msg},
				Gas:  &gas,
			})
			require.NoError(s.T(), err) // The block did not fail.
			if tc.expectedError != nil {
				require.NotEqual(s.T(), uint32(0), res.Code)               // The tx failed.
				require.Contains(s.T(), res.Log, tc.expectedError.Error()) // The tx failed with the expected error.
			} else {
				require.Equal(s.T(), uint32(0), res.Code, res.Log)
			}
		})
	}
}
