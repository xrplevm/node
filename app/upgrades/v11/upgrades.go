package v11

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/log"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	icahosttypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/host/types"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	"github.com/ethereum/go-ethereum/common"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	icaHostKeeper ICAHostKeeper,
	stakingKeeper StakingKeeper,
	transferKeeper TransferKeeper,
	evmKeeper EvmKeeper,
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

		logger.Info("Withdrawing Elys escrow to provided address...")
		if err := withdrawElysEscrow(ctx, logger, transferKeeper); err != nil {
			return nil, err
		}

		logger.Info("Installing missing default preinstalls...")
		if err := installMissingPreinstalls(ctx, logger, evmKeeper); err != nil {
			return nil, err
		}

		logger.Info("Finished v11 upgrade handler")
		return vm, nil
	}
}

// installMissingPreinstalls deploys any default preinstall that is absent.
func installMissingPreinstalls(ctx sdk.Context, logger log.Logger, evmKeeper EvmKeeper) error {
	var missingPreinstalls []evmtypes.Preinstall
	for _, aDefaultPreinstall := range evmtypes.DefaultPreinstalls {
		if evmKeeper.IsContract(ctx, common.HexToAddress(aDefaultPreinstall.Address)) {
			continue
		}
		logger.Info("installing missing preinstall", "name", aDefaultPreinstall.Name, "address", aDefaultPreinstall.Address)
		missingPreinstalls = append(missingPreinstalls, aDefaultPreinstall)
	}
	if len(missingPreinstalls) == 0 {
		return nil
	}
	return evmKeeper.AddPreinstalls(ctx, missingPreinstalls)
}

// withdrawElysEscrow releases the configured amount of XRP from the Elys channel
// escrow to the recovery address configured for the running network.
func withdrawElysEscrow(ctx sdk.Context, logger log.Logger, transferKeeper TransferKeeper) error {
	recoveryCfg, ok := ElysRecoveryByNetwork[ctx.ChainID()]
	if !ok || recoveryCfg.ChannelID == "" {
		logger.Info("no Elys escrow recovery configured for this network, skipping", "chainID", ctx.ChainID())
		return nil
	}

	if !recoveryCfg.Coin.IsPositive() {
		logger.Info("Elys escrow recovery amount is zero, nothing to withdraw", "chainID", ctx.ChainID())
		return nil
	}

	escrowAddr := transfertypes.GetEscrowAddress(transfertypes.PortID, recoveryCfg.ChannelID)

	destAddr, err := sdk.AccAddressFromBech32(recoveryCfg.WithdrawalAddress)
	if err != nil {
		return fmt.Errorf("invalid withdrawal address %q: %w", recoveryCfg.WithdrawalAddress, err)
	}

	if err := transferKeeper.UnescrowCoin(ctx, escrowAddr, destAddr, recoveryCfg.Coin); err != nil {
		return fmt.Errorf("failed to unescrow %s from elys escrow: %w", recoveryCfg.Coin, err)
	}

	logger.Info("Withdrew stranded Elys escrow", "amount", recoveryCfg.Coin.String(), "from", escrowAddr.String(), "to", destAddr.String())
	return nil
}
