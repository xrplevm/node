package v9

import (
	"context"
	"errors"
	"strconv"

	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	erc20types "github.com/cosmos/evm/x/erc20/types"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/ethereum/go-ethereum/common"
	legacyevmtypes "github.com/xrplevm/node/v9/legacy/evm/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	storeKeys map[string]*storetypes.KVStoreKey,
	appCodec codec.Codec,
	evmKeeper EvmKeeper,
	erc20Keeper ERC20Keeper,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(c)
		logger := ctx.Logger().With("upgrade", UpgradeName)
		logger.Info("Running v9 upgrade handler...")

		migrateErc20Module(
			ctx,
			storeKeys,
			erc20Keeper,
		)
		migrateEvmModule(
			ctx,
			storeKeys,
			appCodec,
			evmKeeper,
		)

		logger.Info("Finished v9 upgrade handler")

		return mm.RunMigrations(ctx, configurator, vm)
	}
}

func migrateEvmModule(ctx sdk.Context, keys map[string]*storetypes.KVStoreKey, codec codec.Codec, evmKeeper EvmKeeper) error {
	store := ctx.KVStore(keys[evmtypes.StoreKey])

	legacyBz := store.Get(legacyevmtypes.KeyPrefixParams)
	if legacyBz == nil {
		return errors.New("legacyBz cannot be nil")
	}
	var legacyEvmParams legacyevmtypes.Params

	codec.MustUnmarshal(legacyBz, &legacyEvmParams)

	eips := make([]int64, 0, len(legacyEvmParams.ExtraEIPs))

	for i, extraEIP := range legacyEvmParams.ExtraEIPs {
		intEIP, err := strconv.ParseInt(extraEIP, 10, 64)
		if err != nil {
			return err
		}
		eips[i] = intEIP
	}

	accessControl := evmtypes.AccessControl{
		Create: evmtypes.AccessControlType{
			AccessType:        evmtypes.AccessType(legacyEvmParams.AccessControl.Create.AccessType),
			AccessControlList: legacyEvmParams.AccessControl.Create.AccessControlList,
		},
		Call: evmtypes.AccessControlType{
			AccessType:        evmtypes.AccessType(legacyEvmParams.AccessControl.Call.AccessType),
			AccessControlList: legacyEvmParams.AccessControl.Call.AccessControlList,
		},
	}

	params := evmtypes.Params{
		EvmDenom:                legacyEvmParams.EvmDenom,
		ExtraEIPs:               eips,
		AllowUnprotectedTxs:     legacyEvmParams.AllowUnprotectedTxs,
		EVMChannels:             legacyEvmParams.EVMChannels,
		AccessControl:           accessControl,
		ActiveStaticPrecompiles: legacyEvmParams.ActiveStaticPrecompiles,
	}

	return evmKeeper.SetParams(ctx, params)
}

func migrateErc20Module(ctx sdk.Context, keys map[string]*storetypes.KVStoreKey, keeper ERC20Keeper) {
	// In your upgrade handler
	store := ctx.KVStore(keys[erc20types.StoreKey])
	const addressLength = 42 // "0x" + 40 hex characters

	// Migrate dynamic precompiles (IBC tokens, token factory)
	if oldData := store.Get([]byte("DynamicPrecompiles")); len(oldData) > 0 {
		for i := 0; i < len(oldData); i += addressLength {
			address := common.HexToAddress(string(oldData[i : i+addressLength]))
			keeper.SetDynamicPrecompile(ctx, address)
		}
		store.Delete([]byte("DynamicPrecompiles"))
	}

	// Migrate native precompiles
	if oldData := store.Get([]byte("NativePrecompiles")); len(oldData) > 0 {
		for i := 0; i < len(oldData); i += addressLength {
			address := common.HexToAddress(string(oldData[i : i+addressLength]))
			keeper.SetNativePrecompile(ctx, address)
		}
		store.Delete([]byte("NativePrecompiles"))
	}
}
