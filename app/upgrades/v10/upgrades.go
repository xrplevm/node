package v10

import (
	"context"

	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	v9 "github.com/xrplevm/node/v9/app/upgrades/v9"
)

const MainnetChainID = "xrplevm_1440000-1"

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	storeKeys map[string]*storetypes.KVStoreKey,
	appCodec codec.Codec,
	accountKeeper authkeeper.AccountKeeper,
	evmKeeper EvmKeeper,
	erc20Keeper ERC20Keeper,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(c)
		logger := ctx.Logger().With("upgrade", UpgradeName)

		// Also run v9 upgrade handler for mainnet
		if ctx.ChainID() == MainnetChainID {
			logger.Info("Detected mainnet chain id falling back to v9 upgrade handler...")
			err := v9.UpgradeHandler(ctx, storeKeys, appCodec, accountKeeper, evmKeeper, erc20Keeper)
			if err != nil {
				return nil, err
			}
		}
		ctx.Logger().Info("Running v10 upgrade handler...")
		ctx.Logger().Info("Init evm coin info...")
		if err := evmKeeper.InitEvmCoinInfo(ctx); err != nil {
			return nil, err
		}
		ctx.Logger().Info("Finished v10 upgrade handler")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}
