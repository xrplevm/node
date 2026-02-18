package v10

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	evmKeeper EvmKeeper,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(c)
		ctx.Logger().Info("Running v10 upgrade handler...")

		ctx.Logger().Info("Init evm coin info...")
		if err := evmKeeper.InitEvmCoinInfo(ctx); err != nil {
			return nil, err
		}

		ctx.Logger().Info("Finished v10 upgrade handler")

		return mm.RunMigrations(ctx, configurator, vm)
	}
}
