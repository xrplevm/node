package integration

import (
	"time"

	"cosmossdk.io/math"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/stretchr/testify/require"
	"github.com/xrplevm/node/v6/testutil/integration/exrp/utils"
)

// Slashing tests

func (s *TestSuite) TestSlashing_ChangeParams() {
	tt := []struct {
		name          string
		newParams     slashingtypes.Params
		expectedError string
	}{
		{
			name: "change slashing params - invalid slash fraction double sign",
			newParams: slashingtypes.NewParams(
				200,
				math.LegacyOneDec(),
				time.Second,
				math.LegacyZeroDec(),
				math.LegacyOneDec(),
			),
			expectedError: "downtime sign slash fraction must be zero: 1.000000000000000000",
		},
		{
			name: "change slashing params - invalid slash fraction downtime",
			newParams: slashingtypes.NewParams(
				200,
				math.LegacyOneDec(),
				time.Second,
				math.LegacyOneDec(),
				math.LegacyZeroDec(),
			),
			expectedError: "double sign slash fraction must be zero: 1.000000000000000000",
		},
		{
			name: "change slashing params - success",
			newParams: slashingtypes.NewParams(
				200,
				math.LegacyOneDec(),
				time.Second,
				math.LegacyZeroDec(),
				math.LegacyZeroDec(),
			),
		},
	}

	for _, tc := range tt {
		s.Run(tc.name, func() {
			authority := sdktypes.AccAddress(address.Module("gov"))
			msg := &slashingtypes.MsgUpdateParams{
				Authority: authority.String(),
				Params:    tc.newParams,
			}

			proposal, err := utils.SubmitAndAwaitProposalResolution(s.factory, s.Network(), s.keyring.GetKeys(), "test", msg)
			require.NoError(s.T(), err)

			if tc.expectedError != "" {
				require.Equal(s.T(), govv1.ProposalStatus_PROPOSAL_STATUS_FAILED, proposal.Status)
				require.Equal(s.T(), proposal.FailedReason, tc.expectedError)
			} else {
				require.Equal(s.T(), govv1.ProposalStatus_PROPOSAL_STATUS_PASSED, proposal.Status)
			}
		})
	}
}
