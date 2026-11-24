//go:build test

package integration

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/evm/testutil/integration/base/factory"
	"github.com/stretchr/testify/require"
	"github.com/xrplevm/node/v10/testutil/integration/exrp/utils"
)

// Legacy content proposals have no handler, so the submission fails.
func (s *TestSuite) TestGov_LegacyProposal() {
	tt := []struct {
		name    string
		content govv1beta1.Content
	}{
		{
			name:    "text",
			content: govv1beta1.NewTextProposal("text", "text"),
		},
		{
			name: "parameter change",
			content: paramproposal.NewParameterChangeProposal("params", "params", []paramproposal.ParamChange{
				paramproposal.NewParamChange(slashingtypes.ModuleName, string(slashingtypes.KeySignedBlocksWindow), `"200"`),
			}),
		},
	}

	for _, tc := range tt {
		s.Run(tc.name, func() {
			t := s.T()
			key := s.keyring.GetKey(0)
			deposit := sdk.NewCoins(sdk.NewCoin(s.network.GetBaseDenom(), s.network.GetMinDepositAmt()))
			gas := uint64(1_000_000)

			msg, err := govv1beta1.NewMsgSubmitProposal(tc.content, deposit, key.AccAddr)
			require.NoError(t, err)
			res, err := s.factory.CommitCosmosTx(key.Priv, factory.CosmosTxArgs{Msgs: []sdk.Msg{msg}, Gas: &gas})
			require.NoError(t, err)
			require.Equal(t, govtypes.ErrNoProposalHandlerExists.Codespace(), res.Codespace, res.Log)
			require.Equal(t, govtypes.ErrNoProposalHandlerExists.ABCICode(), res.Code, res.Log)
		})
	}
}

// A message proposal replaces the legacy parameter change: it passes and
// changes the params.
func (s *TestSuite) TestGov_MessageProposal() {
	t := s.T()

	params, err := s.network.GetSlashingClient().Params(s.network.GetContext(), &slashingtypes.QueryParamsRequest{})
	require.NoError(t, err)
	updated := params.Params
	updated.SignedBlocksWindow++

	proposal, err := utils.SubmitAndAwaitProposalResolution(s.factory, s.Network(), s.keyring.GetKeys(), "params", &slashingtypes.MsgUpdateParams{
		Authority: sdk.AccAddress(address.Module("gov")).String(),
		Params:    updated,
	})
	require.NoError(t, err)
	require.Equal(t, govv1.ProposalStatus_PROPOSAL_STATUS_PASSED, proposal.Status, proposal.FailedReason)

	params, err = s.network.GetSlashingClient().Params(s.network.GetContext(), &slashingtypes.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, updated, params.Params)
}
