// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package exrpupgrade

import (
	"fmt"
	"os"

	"github.com/xrplevm/node/v4/app"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/gogoproto/proto"

	"cosmossdk.io/log"
	dbm "github.com/cosmos/cosmos-db"
	simutils "github.com/cosmos/cosmos-sdk/testutil/sims"
	erc20types "github.com/evmos/evmos/v20/x/erc20/types"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
	feemarkettypes "github.com/evmos/evmos/v20/x/feemarket/types"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
)

func MustGetIntegrationTestNodeHome() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	return wd + "/../../"
}

// genSetupFn is the type for the module genesis setup functions
type genSetupFn func(evmosApp *app.App, genesisState app.GenesisState, customGenesis interface{}) (app.GenesisState, error)


var genesisSetupFunctions = map[string]genSetupFn{
	evmtypes.ModuleName:       genStateSetter[*evmtypes.GenesisState](evmtypes.ModuleName),
	erc20types.ModuleName:     genStateSetter[*erc20types.GenesisState](erc20types.ModuleName),
	govtypes.ModuleName:       genStateSetter[*govtypesv1.GenesisState](govtypes.ModuleName),
	feemarkettypes.ModuleName: genStateSetter[*feemarkettypes.GenesisState](feemarkettypes.ModuleName),
	distrtypes.ModuleName:     genStateSetter[*distrtypes.GenesisState](distrtypes.ModuleName),
	banktypes.ModuleName:      genStateSetter[*banktypes.GenesisState](banktypes.ModuleName),
	authtypes.ModuleName:      genStateSetter[*authtypes.GenesisState](authtypes.ModuleName),
	capabilitytypes.ModuleName: genStateSetter[*capabilitytypes.GenesisState](capabilitytypes.ModuleName),
	genutiltypes.ModuleName:  genStateSetter[*genutiltypes.GenesisState](genutiltypes.ModuleName),
}

// genStateSetter is a generic function to set module-specific genesis state
func genStateSetter[T proto.Message](moduleName string) genSetupFn {
	return func(exrpApp *app.App, genesisState app.GenesisState, customGenesis interface{}) (app.GenesisState, error) {
		var customGen T
		err := exrpApp.AppCodec().UnmarshalJSON(genesisState[moduleName], customGen)
		
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling %s module genesis state: %w", moduleName, err)
		}
		// moduleGenesis, ok := customGenesis.(T)
		// if !ok {
		// 	return nil, fmt.Errorf("invalid type %T for %s module genesis state", customGenesis, moduleName)
		// }

		genesisState[moduleName] = exrpApp.AppCodec().MustMarshalJSON(customGen)
		return genesisState, nil
	}
}

// customizeGenesis modifies genesis state if there're any custom genesis state
// for specific modules
func customizeGenesis(exrpApp *app.App, customGen CustomGenesisState, genesisState app.GenesisState) (app.GenesisState, error) {
	var err error
	for mod, modGenState := range customGen {
		if fn, found := genesisSetupFunctions[mod]; found {
			genesisState, err = fn(exrpApp, genesisState, modGenState)
			if err != nil {
				return genesisState, err
			}
		} else {
			panic(fmt.Sprintf("module %s not found in genesis setup functions", mod))
		}
	}
	return genesisState, err
}

// createExrpApp creates an exrp app
func createExrpApp(chainID string, customBaseAppOptions ...func(*baseapp.BaseApp)) *app.App {
	testNodeHome := MustGetIntegrationTestNodeHome()
	// Create evmos app
	db := dbm.NewMemDB()
	logger := log.NewNopLogger()
	loadLatest := true
	skipUpgradeHeights := map[int64]bool{}
	homePath := testNodeHome
	invCheckPeriod := uint(5)
	appOptions := simutils.NewAppOptionsWithFlagHome(homePath)
	baseAppOptions := append(customBaseAppOptions, baseapp.SetChainID(chainID)) //nolint:gocritic

	return app.New(
		logger,
		db,
		nil,
		loadLatest,
		skipUpgradeHeights,
		homePath,
		invCheckPeriod,
		appOptions,
		baseAppOptions...,
	)
}
