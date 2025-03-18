package app

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/evmos/evmos/v20/app/ante"
	ethante "github.com/evmos/evmos/v20/app/ante/evm"
	etherminttypes "github.com/evmos/evmos/v20/types"
	poaante "github.com/xrplevm/node/v6/x/poa/ante"
)

type AnteHandlerOptions ante.HandlerOptions

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
		TxFeeChecker:           ethante.NewDynamicFeeChecker(app.EvmKeeper),
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
