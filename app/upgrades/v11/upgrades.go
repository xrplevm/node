package v11

import (
	"context"
	"fmt"

	"cosmossdk.io/log"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	icahosttypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/host/types"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	icaHostKeeper ICAHostKeeper,
	bankKeeper BankKeeper,
	transferKeeper TransferKeeper,
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

		logger.Info("Disabling ICA host module...")
		icaHostKeeper.SetParams(ctx, icahosttypes.NewParams(false, nil))

		logger.Info("Withdrawing Elys escrow to provided address...")
		if err := withdrawElysEscrow(ctx, logger, bankKeeper, transferKeeper); err != nil {
			return nil, err
		}

		logger.Info("Finished v11 upgrade handler")
		return vm, nil
	}
}

// withdrawElysEscrow moves the XRP stranded in the Elys transfer-channel
// escrow to the recovery address and updates the total-escrow accounting.
func withdrawElysEscrow(ctx sdk.Context, logger log.Logger, bankKeeper BankKeeper, transferKeeper TransferKeeper) error {
	escrowAddr := transfertypes.GetEscrowAddress(transfertypes.PortID, ElysChannelID)
	if escrowAddr.String() != ElysEscrowAddress {
		return fmt.Errorf("elys escrow mismatch: derived %s for %s/%s, expected %s",
			escrowAddr, transfertypes.PortID, ElysChannelID, ElysEscrowAddress)
	}

	destAddr, err := sdk.AccAddressFromBech32(WithdrawalAddress)
	if err != nil {
		return fmt.Errorf("invalid withdrawal address %q: %w", WithdrawalAddress, err)
	}

	// XRP base denom from the canonical EVM coin config.
	xrpDenom := evmtypes.GetEVMCoinDenom()
	escrowBalance := bankKeeper.GetBalance(ctx, escrowAddr, xrpDenom)
	if escrowBalance.IsZero() {
		logger.Info("Elys escrow already empty, nothing to withdraw", "escrow", escrowAddr.String())
		return nil
	}

	if err := bankKeeper.SendCoins(ctx, escrowAddr, destAddr, sdk.NewCoins(escrowBalance)); err != nil {
		return fmt.Errorf("failed to withdraw elys escrow: %w", err)
	}

	totalEscrow := transferKeeper.GetTotalEscrowForDenom(ctx, xrpDenom)
	transferKeeper.SetTotalEscrowForDenom(ctx, totalEscrow.Sub(escrowBalance))

	logger.Info("Withdrew stranded Elys XRP", "amount", escrowBalance.String(), "from", escrowAddr.String(), "to", destAddr.String())
	return nil
}
