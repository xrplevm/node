package v9

import (
	"context"

	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	erc20types "github.com/cosmos/evm/x/erc20/types"
	"github.com/ethereum/go-ethereum/common"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	storeKeys map[string]*storetypes.KVStoreKey,
	erc20Keeper ERC20Keeper,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(c)
		logger := ctx.Logger().With("upgrade", UpgradeName)
		logger.Info("Running v9 upgrade handler...")

		// In your upgrade handler
		store := ctx.KVStore(storeKeys[erc20types.StoreKey])
		const addressLength = 42 // "0x" + 40 hex characters

		// Migrate dynamic precompiles (IBC tokens, token factory)
		if oldData := store.Get([]byte("DynamicPrecompiles")); len(oldData) > 0 {
			for i := 0; i < len(oldData); i += addressLength {
				address := common.HexToAddress(string(oldData[i : i+addressLength]))
				erc20Keeper.SetDynamicPrecompile(ctx, address)
			}
			store.Delete([]byte("DynamicPrecompiles"))
		}

		// Migrate native precompiles
		if oldData := store.Get([]byte("NativePrecompiles")); len(oldData) > 0 {
			for i := 0; i < len(oldData); i += addressLength {
				address := common.HexToAddress(string(oldData[i : i+addressLength]))
				erc20Keeper.SetNativePrecompile(ctx, address)
			}
			store.Delete([]byte("NativePrecompiles"))
		}

		logger.Info("Finished v9 upgrade handler")

		return mm.RunMigrations(ctx, configurator, vm)
	}
}
