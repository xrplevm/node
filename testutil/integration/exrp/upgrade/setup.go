package exrpupgrade

import (
	"cosmossdk.io/log"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	simutils "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/xrplevm/node/v5/app"
	exrpcommon "github.com/xrplevm/node/v5/testutil/integration/exrp/common"
)

const (
	DefaultNodeDBName = "application"
	DefaultNodeDBDir  = ".exrpd/data"
)

// createExrpApp creates an exrp app
func CreateExrpApp(chainID string, customBaseAppOptions ...func(*baseapp.BaseApp)) *app.App {
	testNodeHome := exrpcommon.MustGetIntegrationTestNodeHome()
	// Create exrp app
	db, err := dbm.NewGoLevelDB(DefaultNodeDBName, DefaultNodeDBDir, nil)
	if err != nil {
		panic(err)
	}
	logger := log.NewNopLogger()
	loadLatest := false 
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
