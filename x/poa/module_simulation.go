package poa

import (
	"math/rand"

	"github.com/Peersyst/exrp/testutil/sample"
	poasimulation "github.com/Peersyst/exrp/x/poa/simulation"
	"github.com/Peersyst/exrp/x/poa/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// avoid unused import issue
var (
	_ = sample.AccAddress
	_ = poasimulation.FindAccount
	_ = simappparams.StakePerAccount
	_ = simulation.MsgEntryKind
	_ = baseapp.Paramspace
)

const (
// this line is used by starport scaffolding # simapp/module/const
)

// GenerateGenesisState creates a randomized GenState of the module
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	accs := make([]string, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		accs[i] = acc.Address.String()
	}
	poaGenesis := types.GenesisState{
		Params: types.DefaultParams(),
		// this line is used by starport scaffolding # simapp/module/genesisState
	}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&poaGenesis)
}

// ProposalContents doesn't return any content functions for governance proposals
func (am AppModule) ProposalContents(simState module.SimulationState) []simtypes.WeightedProposalContent {
	proposalContents := make([]simtypes.WeightedProposalContent, 0)

	proposalContents = append(proposalContents, poasimulation.AddValidatorProposal())
	// proposalContents = append(proposalContents, poasimulation.RemoveValidatorProposal(am.bk, am.sk))

	return proposalContents
}

// RandomizedParams creates randomized  param changes for the simulator
func (am AppModule) RandomizedParams(_ *rand.Rand) []simtypes.ParamChange {

	return []simtypes.ParamChange{}
}

// RegisterStoreDecoder registers a decoder
func (am AppModule) RegisterStoreDecoder(_ sdk.StoreDecoderRegistry) {}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	operations := make([]simtypes.WeightedOperation, 0)
	// this line is used by starport scaffolding # simapp/module/operation

	return operations
}
