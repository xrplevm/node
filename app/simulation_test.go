package app_test

import (
	"github.com/Peersyst/exrp/app"
	dbm "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	simulationtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	simcli "github.com/cosmos/cosmos-sdk/x/simulation/client/cli"
	"github.com/evmos/evmos/v15/app/ante"
	ethante "github.com/evmos/evmos/v15/app/ante/evm"
	evmostypes "github.com/evmos/evmos/v15/types"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func init() {
	simcli.GetSimulatorFlags()
}

const SimAppChainID = "simulation_777-1"

// fauxMerkleModeOpt returns a BaseApp option to use a dbStoreAdapter instead of
// an IAVLStore for faster simulation speed.
func fauxMerkleModeOpt(bapp *baseapp.BaseApp) {
	bapp.SetFauxMerkleMode()
}

// NewSimApp disable feemarket on native tx, otherwise the cosmos-sdk simulation tests will fail.
func NewSimApp(logger log.Logger, db dbm.DB, config simulationtypes.Config) (*app.App, error) {
	encodingConfig := app.MakeEncodingConfig()
	appOptions := make(simtestutil.AppOptionsMap, 0)
	appOptions[flags.FlagHome] = app.DefaultNodeHome
	appOptions[server.FlagInvCheckPeriod] = simcli.FlagPeriodValue

	bApp := app.New(
		logger,
		db,
		nil,
		false,
		map[int64]bool{},
		app.DefaultNodeHome,
		simcli.FlagPeriodValue, encodingConfig,
		appOptions,
		baseapp.SetChainID(config.ChainID),
	)
	// disable feemarket on native tx
	options := ante.HandlerOptions{
		Cdc:                    encodingConfig.Codec,
		AccountKeeper:          bApp.AccountKeeper,
		BankKeeper:             bApp.BankKeeper,
		ExtensionOptionChecker: evmostypes.HasDynamicFeeExtensionOption,
		EvmKeeper:              bApp.EvmKeeper,
		FeegrantKeeper:         bApp.FeeGrantKeeper,
		IBCKeeper:              bApp.IBCKeeper,
		FeeMarketKeeper:        bApp.FeeMarketKeeper,
		SignModeHandler:        encodingConfig.TxConfig.SignModeHandler(),
		SigGasConsumer:         ante.SigVerificationGasConsumer,
		MaxTxGasWanted:         0,
		TxFeeChecker:           ethante.NewDynamicFeeChecker(bApp.EvmKeeper),
	}

	if err := options.Validate(); err != nil {
		panic(err)
	}

	bApp.SetAnteHandler(ante.NewAnteHandler(options))
	if err := bApp.LoadLatestVersion(); err != nil {
		return nil, err
	}
	return bApp, nil
}

// BenchmarkSimulation run the chain simulation
// Running using starport command:
// `ignite chain simulate -v --numBlocks 200 --blockSize 50`
// Running as go benchmark test:
// `go test -benchmem -run=^$ -bench ^BenchmarkSimulation ./app -NumBlocks=200 -BlockSize 50 -Commit=true -Verbose=true -Enabled=true`
func BenchmarkSimulation(b *testing.B) {
	simcli.FlagSeedValue = time.Now().Unix()
	simcli.FlagVerboseValue = true
	simcli.FlagCommitValue = true
	simcli.FlagEnabledValue = true

	config := simcli.NewConfigFromFlags()
	config.ChainID = SimAppChainID
	db, dir, logger, _, err := simtestutil.SetupSimulation(
		config,
		"leveldb-bApp-sim",
		"Simulation",
		simcli.FlagVerboseValue,
		simcli.FlagEnabledValue,
	)

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
		simtestutil.AppStateFn(
			bApp.AppCodec(),
			bApp.SimulationManager(),
			app.NewDefaultGenesisState(bApp.AppCodec()),
		),
		simulationtypes.RandomAccounts,
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
