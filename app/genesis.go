package app

import (
	"encoding/json"

	sdkmath "cosmossdk.io/math"

	feemarkettypes "github.com/evmos/evmos/v20/x/feemarket/types"
)

// GenesisState The genesis state of the blockchain is represented here as a map of raw json
// messages key'd by a identifier string.
// The identifier is used to determine which module genesis information belongs
// to so it may be appropriately routed during init chain.
// Within this application default genesis information is retrieved from
// the ModuleBasicManager which populates json from each BasicModule
// object provided to it during init.
type GenesisState map[string]json.RawMessage

// NewDefaultGenesisState generates the default state for the application.
func NewDefaultGenesisState(app *App) GenesisState {
	genState := app.BasicModuleManager.DefaultGenesis(app.AppCodec())
	// Set default feemarket params
	var feeMarketState feemarkettypes.GenesisState
	app.cdc.MustUnmarshalJSON(genState[feemarkettypes.ModuleName], &feeMarketState)
	feeMarketState.Params.NoBaseFee = true
	feeMarketState.Params.BaseFee = sdkmath.NewInt(0)
	genState[feemarkettypes.ModuleName] = app.cdc.MustMarshalJSON(&feeMarketState)

	return genState
}
