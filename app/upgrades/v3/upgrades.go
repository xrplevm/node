package v2

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	v16 "github.com/evmos/evmos/v20/app/upgrades/v16"
	bankprecompile "github.com/evmos/evmos/v20/precompiles/bank"
	"github.com/evmos/evmos/v20/precompiles/bech32"
	"github.com/evmos/evmos/v20/precompiles/p256"
	evmkeeper "github.com/evmos/evmos/v20/x/evm/keeper"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v13
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	ek *evmkeeper.Keeper,
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
		bankAddress := bankprecompile.Precompile{}.Address()
		if err := ek.EnableStaticPrecompiles(ctx, p256Address, bech32Address, bankAddress); err != nil {
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
