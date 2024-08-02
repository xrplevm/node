package app

import (
	"fmt"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	v2 "github.com/node/xrplevm/v2/app/upgrades/v2"
)

func (app *App) setupUpgradeHandlers() {
	// !! ATTENTION !!
	// v14 upgrade handler
	// !! WHEN UPGRADING TO SDK v0.47 MAKE SURE TO INCLUDE THIS
	// source: https://github.com/cosmos/cosmos-sdk/blob/release/v0.47.x/UPGRADING.md#xconsensus
	app.UpgradeKeeper.SetUpgradeHandler(
		v2.UpgradeName,
		v2.CreateUpgradeHandler(
			app.mm, app.configurator,
			app.EvmKeeper,
			app.ConsensusParamsKeeper,
			app.IBCKeeper.ClientKeeper,
			app.ParamsKeeper,
			app.appCodec,
		),
	)

	// When a planned update height is reached, the old binary will panic
	// writing on disk the height and name of the update that triggered it
	// This will read that value, and execute the preparations for the upgrade.
	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(fmt.Errorf("failed to read upgrade info from disk: %w", err))
	}

	if app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		return
	}

	var storeUpgrades *storetypes.StoreUpgrades

	switch upgradeInfo.Name {
	case v2.UpgradeName:
		// !! ATTENTION !!
		// !! WHEN UPGRADING TO SDK v0.47 MAKE SURE TO INCLUDE THIS
		// source: https://github.com/cosmos/cosmos-sdk/blob/release/v0.47.x/UPGRADING.md
		storeUpgrades = &storetypes.StoreUpgrades{
			Added: []string{
				consensusparamtypes.StoreKey,
				crisistypes.ModuleName,
			},
			Deleted: []string{},
		}
	}

	if storeUpgrades != nil {
		// configure store loader that checks if version == upgradeHeight and applies store upgrades
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, storeUpgrades))
	}
}
