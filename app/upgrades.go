package app

import (
	"fmt"

	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/feegrant"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	icahosttypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/host/types"
	v12 "github.com/xrplevm/node/v12/app/upgrades/v12"
)

// setupUpgradeHandlers registers only the current upgrade. Delete the previous
// app/upgrades/v<version>/ when adding a new one: its release tag keeps it.
func (app *App) setupUpgradeHandlers() {
	app.UpgradeKeeper.SetUpgradeHandler(
		v12.UpgradeName,
		v12.CreateUpgradeHandler(
			app.mm,
			app.configurator,
			app.EvmKeeper,
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

	if upgradeInfo.Name == v12.UpgradeName {
		// These stores are also unmounted from NewKVStoreKeys, and rootmulti
		// only deletes mounted stores, so these entries do not wipe data at
		// runtime. They document that authz, icahost, feegrant and crisis were
		// retired in v12. The old data stays on disk, unreachable. A future
		// re-add needs an explicit cleanup first: a plain re-mount fails at
		// startup, and an Added entry fails on the retained iavl versions.
		storeUpgrades := &storetypes.StoreUpgrades{
			Deleted: []string{authz.ModuleName, icahosttypes.StoreKey, feegrant.StoreKey, crisistypes.StoreKey},
		}

		// configure store loader that checks if version == upgradeHeight and applies store upgrades
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, storeUpgrades))
	}
}
