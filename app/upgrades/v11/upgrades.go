package v11

import (
	"context"
	"time"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	icahosttypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/host/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	icaHostKeeper ICAHostKeeper,
	stakingKeeper StakingKeeper,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(c)
		logger := ctx.Logger().With("upgrade", UpgradeName)
		logger.Info("Running v11 upgrade handler...")

		// Run migrations first so no module migration can restore ICA host defaults.
		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return nil, err
		}

		stakingParams, err := stakingKeeper.GetParams(c)
		if err != nil {
			return nil, err
		}

		stakingParams.UnbondingTime = 7 * 24 * time.Hour

		if err := stakingKeeper.SetParams(c, stakingParams); err != nil {
			return nil, err
		}

		logger.Info("Disabling ICA host module...")
		icaHostKeeper.SetParams(ctx, icahosttypes.NewParams(false, nil))
		logger.Info("Finished v11 upgrade handler")
		return vm, nil
	}
}
