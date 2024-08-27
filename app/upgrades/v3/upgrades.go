package v2

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	v16 "github.com/evmos/evmos/v19/app/upgrades/v16"
	v192 "github.com/evmos/evmos/v19/app/upgrades/v19_2"
	"github.com/evmos/evmos/v19/precompiles/bech32"
	"github.com/evmos/evmos/v19/precompiles/p256"
	erc20keeper "github.com/evmos/evmos/v19/x/erc20/keeper"
	evmkeeper "github.com/evmos/evmos/v19/x/evm/keeper"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v13
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	ek *evmkeeper.Keeper,
	erc20Keeper erc20keeper.Keeper,
	ak authkeeper.AccountKeeper,
	bk bankkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)
		/** Evmos v16 upgrades (pre-module upgrades) **/
		// Add Burner role to fee collector
		if err := v16.MigrateFeeCollector(ak, ctx); err != nil {
			logger.Error("failed to migrate the fee collector", "error", err.Error())
			return nil, err
		}

		/** Evmos v19 upgrades (pre-module upgrades) **/
		// Add code extensions
		if err := v192.AddCodeToERC20Extensions(ctx, logger, erc20Keeper); err != nil {
			return nil, err
		}

		/** Module upgrades **/
		logger.Debug("running module migrations ...")
		versionMap, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			logger.Error("failed to migrate modules", "error", err.Error())
			return nil, err
		}

		/** Evmos v16 upgrades (post-module upgrades) **/
		// enable secp256r1 and bech32 precompiles
		p256Address := p256.Precompile{}.Address()
		bech32Address := bech32.Precompile{}.Address()
		if err := ek.EnableStaticPrecompiles(ctx, p256Address, bech32Address); err != nil {
			logger.Error("failed to enable precompiles", "error", err.Error())
			return nil, err
		}

		/** Custom migrations **/
		// Disable default EVM Channels
		params := ek.GetParams(ctx)
		params.EVMChannels = []string{}
		if err := ek.SetParams(ctx, params); err != nil {
			logger.Error("failed to remove EVMChannels from evm params", "error", err.Error())
			return nil, err
		}
		// Add XRP bank metadata for ERC20 recognition
		bk.SetDenomMetaData(ctx, banktypes.Metadata{
			Base: "axrp",
			DenomUnits: []*banktypes.DenomUnit{
				{
					Denom:    "axrp",
					Aliases:  []string{"attoxrp"},
					Exponent: 0,
				},
				{
					Denom:    "xrp",
					Aliases:  []string{},
					Exponent: 18,
				},
			},
			Description: "The native currency of the XRP Ledger",
			Display:     "xrp",
			Name:        "XRP",
			Symbol:      "XRP",
		})

		return versionMap, nil
	}
}
