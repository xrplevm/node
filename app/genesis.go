package app

import (
	"encoding/json"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	erc20types "github.com/cosmos/evm/x/erc20/types"
	evmtypes "github.com/cosmos/evm/x/vm/types"

	feemarkettypes "github.com/cosmos/evm/x/feemarket/types"
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
	genesis := app.BasicModuleManager.DefaultGenesis(app.appCodec)
	evmGenState := NewEVMGenesisState()
	genesis[evmtypes.ModuleName] = app.appCodec.MustMarshalJSON(evmGenState)

	erc20GenState := NewErc20GenesisState()
	genesis[erc20types.ModuleName] = app.appCodec.MustMarshalJSON(erc20GenState)

	feemarketGenState := NewFeeMarketGenesisState()
	genesis[feemarkettypes.ModuleName] = app.appCodec.MustMarshalJSON(feemarketGenState)

	bankGenState := NewBankGenesisState()
	genesis[banktypes.ModuleName] = app.appCodec.MustMarshalJSON(bankGenState)

	return genesis
}

// NewEVMGenesisState returns the default genesis state for the EVM module.
//
// NOTE: for the example chain implementation we need to set the default EVM denomination,
// enable ALL precompiles, and include default preinstalls.
func NewEVMGenesisState() *evmtypes.GenesisState {
	evmGenState := evmtypes.DefaultGenesisState()
	evmGenState.Params.EvmDenom = BaseDenom
	evmGenState.Params.ActiveStaticPrecompiles = evmtypes.AvailableStaticPrecompiles
	evmGenState.Preinstalls = evmtypes.DefaultPreinstalls

	return evmGenState
}

// NewErc20GenesisState returns the default genesis state for the ERC20 module.
//
// NOTE: for the example chain implementation we are also adding a default token pair,
// which is the base denomination of the chain (i.e. the WEVMOS contract).
func NewErc20GenesisState() *erc20types.GenesisState {
	erc20GenState := erc20types.DefaultGenesisState()
	erc20GenState.TokenPairs = []erc20types.TokenPair{
		{
			Erc20Address:  NativeErc20ContractAddress,
			Denom:         BaseDenom,
			Enabled:       true,
			ContractOwner: erc20types.OWNER_MODULE,
		},
	}
	erc20GenState.NativePrecompiles = []string{NativeErc20ContractAddress}

	return erc20GenState
}

// NewFeeMarketGenesisState returns the default genesis state for the feemarket module.
func NewFeeMarketGenesisState() *feemarkettypes.GenesisState {
	return feemarkettypes.DefaultGenesisState()
}

// NewBankGenesisState returns the default genesis state for the bank module.
func NewBankGenesisState() *banktypes.GenesisState {
	bankGenState := banktypes.DefaultGenesisState()
	bankGenState.DenomMetadata = []banktypes.Metadata{
		{
			Description: DenomDescription,
			DenomUnits: []*banktypes.DenomUnit{
				{Denom: BaseDenom, Exponent: 0},
				{Denom: Denom, Exponent: 18},
			},
			Base:    BaseDenom,
			Name:    DenomName,
			Symbol:  DenomSymbol,
			Display: Denom,
		},
	}

	return bankGenState
}
