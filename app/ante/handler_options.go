package ante

import (
	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	ibckeeper "github.com/cosmos/ibc-go/v10/modules/core/keeper"

	txsigning "cosmossdk.io/x/tx/signing"

	anteinterfaces "github.com/cosmos/evm/ante/interfaces"
)

type HandlerOptions struct {
	Cdc                    codec.BinaryCodec
	AccountKeeper          anteinterfaces.AccountKeeper
	BankKeeper             anteinterfaces.BankKeeper
	IBCKeeper              *ibckeeper.Keeper
	FeeMarketKeeper        anteinterfaces.FeeMarketKeeper
	EvmKeeper              anteinterfaces.EVMKeeper
	FeegrantKeeper         ante.FeegrantKeeper
	ExtensionOptionChecker ante.ExtensionOptionChecker
	SignModeHandler        *txsigning.HandlerMap
	SigGasConsumer         func(meter storetypes.GasMeter, sig signing.SignatureV2, params authtypes.Params) error
	MaxTxGasWanted         uint64
	TxFeeChecker           ante.TxFeeChecker
	StakingKeeper          StakingKeeper
	DistributionKeeper     DistributionKeeper
	ExtraDecorator         sdk.AnteDecorator
	AuthzDisabledMsgTypes  []string
}

// Validate checks if the keepers are defined
func (options *HandlerOptions) Validate() error {
	if options.Cdc == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "codec is required for ante builder")
	}
	if options.AccountKeeper == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "account keeper is required for ante builder")
	}
	if options.BankKeeper == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "bank keeper is required for ante builder")
	}
	if options.SignModeHandler == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "sign mode handler is required for ante builder")
	}
	if options.EvmKeeper == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "evm keeper is required for ante builder")
	}
	if options.FeeMarketKeeper == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "feemarket keeper is required for ante builder")
	}
	if options.StakingKeeper == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "staking keeper is required for AnteHandler")
	}
	if options.DistributionKeeper == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "distribution keeper is required for AnteHandler")
	}
	return nil
}
