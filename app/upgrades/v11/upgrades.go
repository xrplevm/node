package v11

import (
	"context"
	"fmt"

	"cosmossdk.io/log"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
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

// withdrawElysEscrow moves every coin stranded in the Elys transfer-channel
// escrow to the recovery address and updates the total-escrow accounting. It
// selects the channel and destination for the running network.
func withdrawElysEscrow(ctx sdk.Context, logger log.Logger, bankKeeper BankKeeper, transferKeeper TransferKeeper) error {
	recovery, ok := ElysRecoveryByNetwork[ctx.ChainID()]
	if !ok || recovery.ChannelID == "" {
		logger.Info("no Elys escrow recovery configured for this network, skipping", "chainID", ctx.ChainID())
		return nil
	}

	escrowAddr := transfertypes.GetEscrowAddress(transfertypes.PortID, recovery.ChannelID)

	destAddr, err := sdk.AccAddressFromBech32(recovery.WithdrawalAddress)
	if err != nil {
		return fmt.Errorf("invalid withdrawal address %q: %w", recovery.WithdrawalAddress, err)
	}

	balances := bankKeeper.GetAllBalances(ctx, escrowAddr)
	if balances.Empty() {
		logger.Info("Elys escrow already empty, nothing to withdraw", "escrow", escrowAddr.String())
		return nil
	}

	if err := bankKeeper.SendCoins(ctx, escrowAddr, destAddr, balances); err != nil {
		return fmt.Errorf("failed to withdraw elys escrow: %w", err)
	}

	// SendCoins bypasses UnescrowCoin, so decrement the per-denom escrow counter
	// ourselves.
	for _, coin := range balances {
		totalEscrow := transferKeeper.GetTotalEscrowForDenom(ctx, coin.Denom)
		// Escrow may hold untracked coins and Sub panics on underflow.
		decrement := coin
		if totalEscrow.IsLT(coin) {
			decrement = totalEscrow
		}
		transferKeeper.SetTotalEscrowForDenom(ctx, totalEscrow.Sub(decrement))
	}

	logger.Info("Withdrew stranded Elys escrow", "amount", balances.String(), "from", escrowAddr.String(), "to", destAddr.String())
	return nil
}
