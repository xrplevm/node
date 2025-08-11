package app

import (
	storetypes "cosmossdk.io/store/types"
	txsigning "cosmossdk.io/x/tx/signing"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/evm/ante"
	ethante "github.com/cosmos/evm/ante/evm"
	anteinterfaces "github.com/cosmos/evm/ante/interfaces"
	etherminttypes "github.com/cosmos/evm/types"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
	poaante "github.com/xrplevm/node/v8/x/poa/ante"
)

type AnteHandlerOptions struct {
	Cdc                    codec.BinaryCodec
	AccountKeeper          anteinterfaces.AccountKeeper
	BankKeeper             anteinterfaces.BankKeeper
	IBCKeeper              *ibckeeper.Keeper
	FeeMarketKeeper        anteinterfaces.FeeMarketKeeper
	EvmKeeper              anteinterfaces.EVMKeeper
	FeegrantKeeper         authante.FeegrantKeeper
	ExtensionOptionChecker authante.ExtensionOptionChecker
	SignModeHandler        *txsigning.HandlerMap
	SigGasConsumer         func(meter storetypes.GasMeter, sig signing.SignatureV2, params authtypes.Params) error
	MaxTxGasWanted         uint64
	TxFeeChecker           authante.TxFeeChecker
	StakingKeeper          StakingKeeper
	DistributionKeeper     DistributionKeeper
	ExtraDecorator         sdk.AnteDecorator
	AuthzDisabledMsgTypes  []string
}

func NewAnteHandlerOptionsFromApp(app *App, txConfig client.TxConfig, maxGasWanted uint64) *AnteHandlerOptions {
	return &AnteHandlerOptions{
		Cdc:                    app.appCodec,
		AccountKeeper:          app.AccountKeeper,
		BankKeeper:             app.BankKeeper,
		ExtensionOptionChecker: etherminttypes.HasDynamicFeeExtensionOption,
		EvmKeeper:              app.EvmKeeper,
		FeegrantKeeper:         app.FeeGrantKeeper,
		IBCKeeper:              app.IBCKeeper,
		FeeMarketKeeper:        app.FeeMarketKeeper,
		SignModeHandler:        txConfig.SignModeHandler(),
		SigGasConsumer:         ante.SigVerificationGasConsumer,
		MaxTxGasWanted:         maxGasWanted,
		TxFeeChecker:           ethante.NewDynamicFeeChecker(app.FeeMarketKeeper),
		StakingKeeper:          app.StakingKeeper,
		DistributionKeeper:     app.DistrKeeper,
		ExtraDecorator:         poaante.NewPoaDecorator(),
		AuthzDisabledMsgTypes: []string{
			sdk.MsgTypeURL(&stakingtypes.MsgUndelegate{}),
			sdk.MsgTypeURL(&stakingtypes.MsgBeginRedelegate{}),
			sdk.MsgTypeURL(&stakingtypes.MsgCancelUnbondingDelegation{}),
			sdk.MsgTypeURL(&stakingtypes.MsgDelegate{}),
		},
	}
}

func (aa *AnteHandlerOptions) Validate() error {
	return (*ante.HandlerOptions)(aa).Validate()
}

func (aa *AnteHandlerOptions) Options() ante.HandlerOptions {
	return ante.HandlerOptions(*aa)
}

func (aa *AnteHandlerOptions) WithCodec(cdc codec.BinaryCodec) *AnteHandlerOptions {
	aa.Cdc = cdc
	return aa
}

func (aa *AnteHandlerOptions) WithMaxTxGasWanted(maxTxGasWanted uint64) *AnteHandlerOptions {
	aa.MaxTxGasWanted = maxTxGasWanted
	return aa
}
