package app_test

import (
	"github.com/Peersyst/exrp/app"
	dbm "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/baseapp"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	simulationtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	simcli "github.com/cosmos/cosmos-sdk/x/simulation/client/cli"
	"github.com/evmos/ethermint/app/ante"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
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

// EmptyAppOptions is a stub implementing AppOptions
type EmptyAppOptions struct{}

// Get implements AppOptions
func (ao EmptyAppOptions) Get(o string) interface{} {
	return nil
}

// NewSimApp disable feemarket on native tx, otherwise the cosmos-sdk simulation tests will fail.
func NewSimApp(logger log.Logger, db dbm.DB) (*app.App, error) {
	encodingConfig := app.MakeEncodingConfig()
	newApp := app.New(logger, db, nil, false, map[int64]bool{}, app.DefaultNodeHome, simcli.FlagPeriodValue, encodingConfig, EmptyAppOptions{}, fauxMerkleModeOpt)
	// disable feemarket on native tx
	anteHandler, err := ante.NewAnteHandler(ante.HandlerOptions{
		AccountKeeper:   newApp.AccountKeeper,
		BankKeeper:      newApp.BankKeeper,
		SignModeHandler: encodingConfig.TxConfig.SignModeHandler(),
		FeegrantKeeper:  newApp.FeeGrantKeeper,
		SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
		IBCKeeper:       newApp.IBCKeeper,
		EvmKeeper:       newApp.EvmKeeper,
		FeeMarketKeeper: newApp.FeeMarketKeeper,
		MaxTxGasWanted:  0,
	})
	if err != nil {
		return nil, err
	}
	newApp.SetAnteHandler(anteHandler)
	if err := newApp.LoadLatestVersion(); err != nil {
		return nil, err
	}
	return newApp, nil
}

// BenchmarkSimulation run the chain simulation
// Running using starport command:
// `ignite chain simulate -v --numBlocks 200 --blockSize 50`
// Running as go benchmark test:
// `go test -benchmem -run=^$ -bench ^BenchmarkSimulation ./app -NumBlocks=200 -BlockSize 50 -Commit=true -Verbose=true -Enabled=true`
func BenchmarkSimulation(b *testing.B) {
	config := simcli.NewConfigFromFlags()
	config.ChainID = SimAppChainID

	db, dir, logger, _, err := simtestutil.SetupSimulation(config, "goleveldb-app-sim", "Simulation", simcli.FlagVerboseValue, simcli.FlagEnabledValue)
	require.NoError(b, err, "simulation setup failed")

	b.Cleanup(func() {
		db.Close()
		err = os.RemoveAll(dir)
		require.NoError(b, err)
	})

	simApp, _ := NewSimApp(logger, db)

	// Run randomized simulations
	_, simParams, simErr := simulation.SimulateFromSeed(
		b,
		os.Stdout,
		simApp.BaseApp,
		AppStateFn(simApp.AppCodec(), simApp.SimulationManager()),
		simulationtypes.RandomAccounts,
		simtestutil.SimulationOperations(simApp, simApp.AppCodec(), config),
		simApp.ModuleAccountAddrs(),
		config,
		simApp.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	err = simtestutil.CheckExportSimulation(simApp, config, simParams)
	require.NoError(b, err)
	require.NoError(b, simErr)

	if config.Commit {
		simtestutil.PrintStats(db)
	}
}
