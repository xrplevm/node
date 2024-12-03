package exrpcommon

import (
	"os"

	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdktypes "github.com/cosmos/cosmos-sdk/types"

	"fmt"

	dbm "github.com/cosmos/cosmos-db"
	simutils "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/gogoproto/proto"
	"github.com/xrplevm/node/v4/app"

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

// genSetupFn is the type for the module genesis setup functions
type genSetupFn func(exrpApp *app.App, genesisState app.GenesisState, customGenesis interface{}) (app.GenesisState, error)

var genesisSetupFunctions = map[string]genSetupFn{
	evmtypes.ModuleName:       GenStateSetter[*evmtypes.GenesisState](evmtypes.ModuleName),
	erc20types.ModuleName:     GenStateSetter[*erc20types.GenesisState](erc20types.ModuleName),
	govtypes.ModuleName:       GenStateSetter[*govtypesv1.GenesisState](govtypes.ModuleName),
	feemarkettypes.ModuleName: GenStateSetter[*feemarkettypes.GenesisState](feemarkettypes.ModuleName),
	distrtypes.ModuleName:     GenStateSetter[*distrtypes.GenesisState](distrtypes.ModuleName),
	banktypes.ModuleName:      GenStateSetter[*banktypes.GenesisState](banktypes.ModuleName),
	authtypes.ModuleName:      GenStateSetter[*authtypes.GenesisState](authtypes.ModuleName),
	capabilitytypes.ModuleName: GenStateSetter[*capabilitytypes.GenesisState](capabilitytypes.ModuleName),
	genutiltypes.ModuleName:  GenStateSetter[*genutiltypes.GenesisState](genutiltypes.ModuleName),
}

// GenStateSetter is a generic function to set module-specific genesis state
func GenStateSetter[T proto.Message](moduleName string) genSetupFn {
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

// CustomizeGenesis modifies genesis state if there're any custom genesis state
// for specific modules
func CustomizeGenesis(exrpApp *app.App, customGen CustomGenesisState, genesisState app.GenesisState) (app.GenesisState, error) {
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


func SetupSdkConfig() {
	accountPubKeyPrefix := accountAddressPrefix + "pub"
	validatorAddressPrefix := accountAddressPrefix + "valoper"
	validatorPubKeyPrefix := accountAddressPrefix + "valoperpub"
	consNodeAddressPrefix := accountAddressPrefix + "valcons"
	consNodePubKeyPrefix := accountAddressPrefix + "valconspub"

	// Set config
	config := sdktypes.GetConfig()
	config.SetBech32PrefixForAccount(accountAddressPrefix, accountPubKeyPrefix)
	config.SetBech32PrefixForValidator(validatorAddressPrefix, validatorPubKeyPrefix)
	config.SetBech32PrefixForConsensusNode(consNodeAddressPrefix, consNodePubKeyPrefix)
	config.SetCoinType(bip44CoinType)
	config.SetPurpose(sdktypes.Purpose) // Shared
	config.Seal()
}

func MustGetIntegrationTestNodeHome() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	return wd + "/../../"
}

// createExrpApp creates an exrp app
func CreateExrpApp(chainID string, customBaseAppOptions ...func(*baseapp.BaseApp)) *app.App {
	testNodeHome := MustGetIntegrationTestNodeHome()
	// Create exrp app
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
