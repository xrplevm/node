package v115

import (
	"context"
	"fmt"
	"slices"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/ethereum/go-ethereum/common"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	evmKeeper EvmKeeper,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(c)
		logger := ctx.Logger().With("upgrade", UpgradeName)
		logger.Info("Running v11.5.0 upgrade handler...")

		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return nil, err
		}

		if err := removeGovAndP256Precompiles(ctx, evmKeeper); err != nil {
			return nil, fmt.Errorf("failed to update active static precompiles: %w", err)
		}

		return vm, nil
	}
}

// removeGovAndP256Precompiles removes the gov precompile and the legacy p256
// static precompile (Osaka enables P256VERIFY natively at 0x...0100) from the
// active static precompiles, keeping all other EVM params.
func removeGovAndP256Precompiles(ctx sdk.Context, evmKeeper EvmKeeper) error {
	gov := common.HexToAddress(evmtypes.GovPrecompileAddress)
	p256 := common.BytesToAddress([]byte{0x01, 0x00})
	params := evmKeeper.GetParams(ctx)
	params.ActiveStaticPrecompiles = slices.DeleteFunc(params.ActiveStaticPrecompiles, func(address string) bool {
		if !common.IsHexAddress(address) {
			return false
		}
		addr := common.HexToAddress(address)
		return addr == gov || addr == p256
	})
	return evmKeeper.SetParams(ctx, params)
}
