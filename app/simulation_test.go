package app_test

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"slices"
	"testing"

	"cosmossdk.io/log"
	"github.com/xrplevm/node/v12/cmd/exrpd/cmd"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/evm/crypto/ethsecp256k1"
	feemarkettypes "github.com/cosmos/evm/x/feemarket/types"
	"github.com/xrplevm/node/v12/app"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	simulationtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	simcli "github.com/cosmos/cosmos-sdk/x/simulation/client/cli"
	"github.com/stretchr/testify/require"
)

func init() {
	simcli.GetSimulatorFlags()
}

const (
	SimAppChainID    = "simulation_777-1"
	SimAppEVMChainID = 777
)

// NewSimApp creates the app used by the simulation tests.
func NewSimApp(logger log.Logger, db dbm.DB, config simulationtypes.Config) (*app.App, error) {
	appOptions := make(simtestutil.AppOptionsMap, 0)
	appOptions[flags.FlagHome] = app.DefaultNodeHome

	evmChainID, err := cmd.CosmosChainIDToEvmChainID(config.ChainID)
	if err != nil {
		return nil, err
	}

	bApp := app.New(
		logger,
		db,
		nil,
		false,
		map[int64]bool{},
		evmChainID,
		appOptions,
		baseapp.SetChainID(config.ChainID),
	)
	if err := bApp.LoadLatestVersion(); err != nil {
		return nil, err
	}
	return bApp, nil
}

// RandomAccounts generates n random accounts
func RandomAccounts(r *rand.Rand, n int) []simulationtypes.Account {
	accs := make([]simulationtypes.Account, n)

	for i := 0; i < n; i++ {
		// don't need that much entropy for simulation
		privkeySeed := make([]byte, 32)
		r.Read(privkeySeed)

		accs[i].PrivKey = &ethsecp256k1.PrivKey{Key: privkeySeed}
		accs[i].PubKey = accs[i].PrivKey.PubKey()
		accs[i].Address = sdk.AccAddress(accs[i].PubKey.Address())

		accs[i].ConsKey = ed25519.GenPrivKeyFromSecret(privkeySeed)
	}

	return accs
}

// simAppStateFn builds the simulation genesis from the app default genesis and adjusts it for the
// simulator. The changes apply only to the simulation. The real node genesis does not change.
func simAppStateFn(bApp *app.App) simulationtypes.AppStateFn {
	cdc := bApp.AppCodec()
	return simtestutil.AppStateFnWithExtendedCbs(
		cdc,
		bApp.SimulationManager(),
		app.NewDefaultGenesisState(bApp),
		// The SDK bank RandomizedGenState replaces the full bank section and drops the axrp denom
		// metadata. x/vm InitGenesis needs it to load the EVM coin info, else InitChain panics with
		// "denom metadata axrp could not be found". Add it back only when it is missing, so metadata
		// from a -Genesis file is kept.
		func(moduleName string, genesisState any) {
			if moduleName != banktypes.ModuleName {
				return
			}
			bankGenState, ok := genesisState.(*banktypes.GenesisState)
			if !ok {
				panic(fmt.Sprintf("unexpected bank genesis type %T", genesisState))
			}
			hasEvmDenom := slices.ContainsFunc(bankGenState.DenomMetadata, func(m banktypes.Metadata) bool {
				return m.Base == app.BaseDenom
			})
			if !hasEvmDenom {
				bankGenState.DenomMetadata = append(bankGenState.DenomMetadata, app.NewBankGenesisState().DenomMetadata...)
			}
		},
		// Simulated txs pay random fees, often 0. With the default 1 gwei base fee, the ante handler
		// rejects them ("gas prices too low"). Both fields are necessary: the evm v0.5+ FeeChecker
		// reads BaseFee directly and ignores NoBaseFee. This callback runs on the final app state,
		// so it also applies to a -Genesis file, and the other feemarket params stay unchanged.
		func(rawState map[string]json.RawMessage) {
			var feemarketGenState feemarkettypes.GenesisState
			cdc.MustUnmarshalJSON(rawState[feemarkettypes.ModuleName], &feemarketGenState)
			feemarketGenState.Params.NoBaseFee = true
			feemarketGenState.Params.BaseFee = sdkmath.LegacyZeroDec()
			rawState[feemarkettypes.ModuleName] = cdc.MustMarshalJSON(&feemarketGenState)
		},
	)
}

// BenchmarkSimulation run the chain simulation
// Running using starport command:
// `ignite chain simulate -v --numBlocks 200 --blockSize 50`
// Running as go benchmark test:
// `go test -benchmem -run=^$ -bench ^BenchmarkSimulation ./app -NumBlocks=200 -BlockSize 50 -Commit=true -Verbose=true -Enabled=true`
//
//nolint:dupl
func BenchmarkSimulation(b *testing.B) {
	simcli.FlagVerboseValue = false
	simcli.FlagCommitValue = true

	config := simcli.NewConfigFromFlags()
	config.ChainID = SimAppChainID
	db, dir, logger, skip, err := simtestutil.SetupSimulation(
		config,
		"leveldb-bApp-sim",
		"Simulation",
		simcli.FlagVerboseValue,
		simcli.FlagEnabledValue,
	)
	if skip {
		b.Skip("skipping application simulation: pass -Enabled=true")
	}

	require.NoError(b, err, "simulation setup failed")

	config.ChainID = SimAppChainID

	b.Cleanup(func() {
		require.NoError(b, db.Close())
		require.NoError(b, os.RemoveAll(dir))
	})

	bApp, _ := NewSimApp(logger, db, config)

	// Run randomized simulations
	_, simParams, simErr := simulation.SimulateFromSeed(
		b,
		os.Stdout,
		bApp.BaseApp,
		simAppStateFn(bApp),
		RandomAccounts,
		simtestutil.SimulationOperations(bApp, bApp.AppCodec(), config),
		bApp.ModuleAccountAddrs(),
		config,
		bApp.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	err = simtestutil.CheckExportSimulation(bApp, config, simParams)
	require.NoError(b, err)
	require.NoError(b, simErr)

	if config.Commit {
		simtestutil.PrintStats(db)
	}
}

//nolint:dupl
func TestFullAppSimulation(t *testing.T) {
	simcli.FlagVerboseValue = false
	simcli.FlagCommitValue = true

	config := simcli.NewConfigFromFlags()
	config.ChainID = SimAppChainID
	db, dir, logger, skip, err := simtestutil.SetupSimulation(
		config,
		"leveldb-bApp-sim",
		"Simulation",
		simcli.FlagVerboseValue,
		simcli.FlagEnabledValue,
	)
	if skip {
		t.Skip("skipping application simulation: pass -Enabled=true")
	}

	require.NoError(t, err, "simulation setup failed")

	config.ChainID = SimAppChainID

	t.Cleanup(func() {
		require.NoError(t, db.Close())
		require.NoError(t, os.RemoveAll(dir))
	})

	bApp, _ := NewSimApp(logger, db, config)

	// Run randomized simulations
	_, simParams, simErr := simulation.SimulateFromSeed(
		t,
		os.Stdout,
		bApp.BaseApp,
		simAppStateFn(bApp),
		RandomAccounts,
		simtestutil.SimulationOperations(bApp, bApp.AppCodec(), config),
		bApp.ModuleAccountAddrs(),
		config,
		bApp.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	err = simtestutil.CheckExportSimulation(bApp, config, simParams)
	require.NoError(t, err)
	require.NoError(t, simErr)

	if config.Commit {
		simtestutil.PrintStats(db)
	}
}
