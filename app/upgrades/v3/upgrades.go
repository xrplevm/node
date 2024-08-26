package v2

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/ethereum/go-ethereum/common"
	v16 "github.com/evmos/evmos/v19/app/upgrades/v16"
	v192 "github.com/evmos/evmos/v19/app/upgrades/v19_2"
	"github.com/evmos/evmos/v19/precompiles/bech32"
	"github.com/evmos/evmos/v19/precompiles/p256"
	erc20keeper "github.com/evmos/evmos/v19/x/erc20/keeper"
	erc20types "github.com/evmos/evmos/v19/x/erc20/types"
	evmkeeper "github.com/evmos/evmos/v19/x/evm/keeper"
	stakingkeeper "github.com/evmos/evmos/v19/x/staking/keeper"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v13
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	ek *evmkeeper.Keeper,
	erc20Keeper erc20keeper.Keeper,
	ak authkeeper.AccountKeeper,
	sk *stakingkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		/**
		Evmos v16 upgrades
		*/
		// enable secp256r1 and bech32 precompile on testnet
		p256Address := p256.Precompile{}.Address()
		bech32Address := bech32.Precompile{}.Address()
		if err := ek.EnableStaticPrecompiles(ctx, p256Address, bech32Address); err != nil {
			logger.Error("failed to enable precompiles", "error", err.Error())
			return nil, err
		}
		// Add Burner role to fee collector
		if err := v16.MigrateFeeCollector(ak, ctx); err != nil {
			logger.Error("failed to migrate the fee collector", "error", err.Error())
			return nil, err
		}

		/**
		Evmos v19 upgrades
		*/
		// Register gas token as a token pair in erc20 module
		pair := erc20types.NewTokenPair(
			common.HexToAddress(erc20types.WEVMOSContractMainnet),
			sk.BondDenom(ctx),
			erc20types.OWNER_MODULE,
		)
		erc20Keeper.SetToken(ctx, pair)
		// Add code extensions
		if err := v192.AddCodeToERC20Extensions(ctx, logger, erc20Keeper); err != nil {
			return nil, err
		}

		// Leave modules are as-is to avoid running InitGenesis.
		logger.Debug("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}
