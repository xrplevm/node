package app

import (
	"fmt"

	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	ratelimittypes "github.com/cosmos/ibc-apps/modules/rate-limiting/v8/types"
	icahosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/types"
	v4 "github.com/xrplevm/node/v3/app/upgrades/v4"
)

func (app *App) setupUpgradeHandlers() {
	authAddr := authtypes.NewModuleAddress(govtypes.ModuleName).String()
	app.UpgradeKeeper.SetUpgradeHandler(
		v4.UpgradeName,
		v4.CreateUpgradeHandler(
			app.mm,
			app.configurator,
			app.appCodec,
			app.GetKey("upgrade"),
			app.ConsensusParamsKeeper,
			authAddr,
			app.EvmKeeper,
			app.Erc20Keeper,
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

	var storeUpgrades *storetypes.StoreUpgrades

	//nolint:gocritic
	switch upgradeInfo.Name {
	case v4.UpgradeName:
		storeUpgrades = &storetypes.StoreUpgrades{
			Added: []string{
				icahosttypes.StoreKey,
				ratelimittypes.ModuleName,
			},
			Deleted: []string{},
		}
	}

	if storeUpgrades != nil {
		// configure store loader that checks if version == upgradeHeight and applies store upgrades
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, storeUpgrades))
	}
}
