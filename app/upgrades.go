package app

import (
	"fmt"
	v4 "github.com/xrplevm/node/v3/app/upgrades/v4"
)

func (app *App) setupUpgradeHandlers() {
	// !! ATTENTION !!
	// v14 upgrade handler
	// !! WHEN UPGRADING TO SDK v0.47 MAKE SURE TO INCLUDE THIS
	// source: https://github.com/cosmos/cosmos-sdk/blob/release/v0.47.x/UPGRADING.md#xconsensus
	app.UpgradeKeeper.SetUpgradeHandler(
		v4.UpgradeName,
		v4.CreateUpgradeHandler(
			app.mm,
			app.configurator,
			app.EvmKeeper,
			app.GovKeeper,
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
}
