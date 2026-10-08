package v12

import (
	"context"
	"slices"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	evmtypes "github.com/cosmos/evm/x/vm/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	evmKeeper EvmKeeper,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(c)
		logger := ctx.Logger().With("upgrade", UpgradeName)
		logger.Info("Running v12 upgrade handler...")

		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return nil, err
		}

		if err := disableGovPrecompile(ctx, evmKeeper); err != nil {
			return nil, err
		}

		return vm, nil
	}
}

func disableGovPrecompile(ctx sdk.Context, evmKeeper EvmKeeper) error {
	params := evmKeeper.GetParams(ctx)
	params.ActiveStaticPrecompiles = slices.DeleteFunc(params.ActiveStaticPrecompiles, func(addr string) bool {
		return addr == evmtypes.GovPrecompileAddress
	})
	return evmKeeper.SetParams(ctx, params)
}
