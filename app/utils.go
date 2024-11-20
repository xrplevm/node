package app

import (
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
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
