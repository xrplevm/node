package app

import (
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
)

// RandomGenesisAccounts defines the default RandomGenesisAccountsFn used on the SDK.
// It creates a slice of BaseAccount, ContinuousVestingAccount and DelayedVestingAccount.
func RandomGenesisAccounts(simState *module.SimulationState) authtypes.GenesisAccounts {
	genesisAccs := make(authtypes.GenesisAccounts, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		bacc := authtypes.NewBaseAccountWithAddress(acc.Address)
		genesisAccs[i] = bacc
	}
	return genesisAccs
}

// govSimModule is x/gov for the simulation manager without legacy content
// proposals. The SDK simulation wraps legacy content in MsgExecLegacyContent,
// which the app does not route (see unroutedMsgMethods), so gov would reject
// every such proposal.
type govSimModule struct {
	gov.AppModule
}

func (govSimModule) ProposalContents(module.SimulationState) []simtypes.WeightedProposalContent {
	return nil
}
