package v11

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
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
	stakingKeeper StakingKeeper,
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
		if err := withdrawElysEscrow(ctx, logger, bankKeeper, transferKeeper); err != nil {
			return nil, err
		}

		logger.Info("Finished v11 upgrade handler")
		return vm, nil
	}
}

// withdrawElysEscrow releases the Elys channel escrow to the recovery address
// configured for the running network.
func withdrawElysEscrow(ctx sdk.Context, logger log.Logger, bankKeeper BankKeeper, transferKeeper TransferKeeper) error {
	recoveryCfg, ok := ElysRecoveryByNetwork[ctx.ChainID()]
	if !ok || recoveryCfg.ChannelID == "" {
		logger.Info("no Elys escrow recovery configured for this network, skipping", "chainID", ctx.ChainID())
		return nil
	}

	escrowAddr := transfertypes.GetEscrowAddress(transfertypes.PortID, recoveryCfg.ChannelID)

	destAddr, err := sdk.AccAddressFromBech32(recoveryCfg.WithdrawalAddress)
	if err != nil {
		return fmt.Errorf("invalid withdrawal address %q: %w", recoveryCfg.WithdrawalAddress, err)
	}

	var released sdk.Coins
	var iterErr error
	transferKeeper.IterateTokensInEscrow(ctx, []byte(transfertypes.KeyTotalEscrowPrefix), func(totalEscrowed sdk.Coin) bool {
		// Cap at the Elys escrow balance so UnescrowCoin's subtraction can't
		// underflow the all-channel total.
		escrowBalance := bankKeeper.GetBalance(ctx, escrowAddr, totalEscrowed.Denom)
		coin := sdk.NewCoin(totalEscrowed.Denom, sdkmath.MinInt(totalEscrowed.Amount, escrowBalance.Amount))
		if !coin.IsPositive() {
			return false
		}

		if err := transferKeeper.UnescrowCoin(ctx, escrowAddr, destAddr, coin); err != nil {
			iterErr = fmt.Errorf("failed to unescrow %s from elys escrow: %w", coin, err)
			return true
		}
		released = released.Add(coin)
		return false
	})
	if iterErr != nil {
		return iterErr
	}

	if released.Empty() {
		logger.Info("Elys escrow already empty, nothing to withdraw", "escrow", escrowAddr.String())
		return nil
	}

	logger.Info("Withdrew stranded Elys escrow", "amount", released.String(), "from", escrowAddr.String(), "to", destAddr.String())
	return nil
}
