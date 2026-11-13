package v12

import (
	"context"
	"slices"

	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	evmtypes "github.com/cosmos/evm/x/vm/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	evmKeeper EvmKeeper,
	keys map[string]*storetypes.KVStoreKey,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(c)
		logger := ctx.Logger().With("upgrade", UpgradeName)
		logger.Info("Running v12 upgrade handler...")

		for _, name := range RetiredStoreKeys {
			deleted, err := deleteAllStoreKeys(ctx.KVStore(keys[name]))
			if err != nil {
				return nil, err
			}
			logger.Info("Emptied retired store", "store", name, "keys_deleted", deleted)
		}

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

// deleteAllStoreKeys deletes every key of the store and returns how many it
// deleted. It collects the keys first: a store must not change under its iterator.
func deleteAllStoreKeys(store storetypes.KVStore) (int, error) {
	var keys [][]byte
	itr := store.Iterator(nil, nil)
	for ; itr.Valid(); itr.Next() {
		keys = append(keys, itr.Key())
	}
	if err := itr.Close(); err != nil {
		return 0, err
	}
	for _, key := range keys {
		store.Delete(key)
	}
	return len(keys), nil
}
