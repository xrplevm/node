package app

import (
	"fmt"

	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	v9 "github.com/xrplevm/node/v9/app/upgrades/v9"

	v5 "github.com/xrplevm/node/v9/app/upgrades/v5"
	v6 "github.com/xrplevm/node/v9/app/upgrades/v6"
	v7 "github.com/xrplevm/node/v9/app/upgrades/v7"
	v8 "github.com/xrplevm/node/v9/app/upgrades/v8"
)

func (app *App) setupUpgradeHandlers() {
	app.UpgradeKeeper.SetUpgradeHandler(
		v5.UpgradeName,
		v5.CreateUpgradeHandler(
			app.mm,
			app.configurator,
		),
	)
	app.UpgradeKeeper.SetUpgradeHandler(
		v6.UpgradeName,
		v6.CreateUpgradeHandler(
			app.mm,
			app.configurator,
		),
	)
	app.UpgradeKeeper.SetUpgradeHandler(
		v7.UpgradeName,
		v7.CreateUpgradeHandler(
			app.mm,
			app.configurator,
		),
	)
	app.UpgradeKeeper.SetUpgradeHandler(
		v8.UpgradeName,
		v8.CreateUpgradeHandler(
			app.mm,
			app.configurator,
		),
	)
	app.UpgradeKeeper.SetUpgradeHandler(
		v9.UpgradeName,
		v9.CreateUpgradeHandler(
			app.mm,
			app.configurator,
			app.keys,
			app.Erc20Keeper,
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
	case v5.UpgradeName, v6.UpgradeName:
		// No store upgrades for v5
		storeUpgrades = &storetypes.StoreUpgrades{}
	}

	if storeUpgrades != nil {
		// configure store loader that checks if version == upgradeHeight and applies store upgrades
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, storeUpgrades))
	}
}
