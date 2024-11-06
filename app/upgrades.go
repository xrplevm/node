package app

import (
	"fmt"

	v2 "github.com/xrplevm/node/v3/app/upgrades/v2"
	v3 "github.com/xrplevm/node/v3/app/upgrades/v3"
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
	app.UpgradeKeeper.SetUpgradeHandler(
		v3.UpgradeName,
		v3.CreateUpgradeHandler(
			app.mm, app.configurator,
			app.EvmKeeper,
			app.AccountKeeper,
			app.BankKeeper,
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
