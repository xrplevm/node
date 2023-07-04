package simulation

import (
	"github.com/Peersyst/exrp/x/poa/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

const (
	OpWeightSubmitAddValidatorProposal = "op_weight_submit_add_validator_proposal"
	DefaultWeightAddValidatorProposal  = 100

	OpWeightSubmitRemoveValidatorProposal = "op_weight_submit_remove_validator_proposal"
	DefaultWeightRemoveValidatorProposal  = 1
)

func AddValidatorProposal() simtypes.WeightedProposalContent {
	return simulation.NewWeightedProposalContent(
		OpWeightSubmitAddValidatorProposal,
		DefaultWeightAddValidatorProposal,
		SimulateAddValidatorProposalContent(),
	)
}

func RemoveValidatorProposal(
	bk types.BankKeeper,
	sk types.StakingKeeper,
) simtypes.WeightedProposalContent {
	return simulation.NewWeightedProposalContent(
		OpWeightSubmitRemoveValidatorProposal,
		DefaultWeightRemoveValidatorProposal,
		SimulateRemoveValidatorProposalContent(bk, sk),
	)
}
