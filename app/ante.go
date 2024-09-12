package app

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/evmos/evmos/v19/app/ante"
	ethante "github.com/evmos/evmos/v19/app/ante/evm"
	etherminttypes "github.com/evmos/evmos/v19/types"
	poaante "github.com/xrplevm/node/v3/x/poa/ante"

	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

type AppAnteHandlerOptions ante.HandlerOptions

func NewAppAnteHandlerOptionsFromApp(app *App) *AppAnteHandlerOptions {
	return &AppAnteHandlerOptions{
		Cdc:                    app.appCodec,
		AccountKeeper:          app.AccountKeeper,
		BankKeeper:             app.BankKeeper,
		ExtensionOptionChecker: etherminttypes.HasDynamicFeeExtensionOption,
		EvmKeeper:              app.EvmKeeper,
		FeegrantKeeper:         app.FeeGrantKeeper,
		IBCKeeper:              app.IBCKeeper,
		FeeMarketKeeper:        app.FeeMarketKeeper,
		SignModeHandler:        nil,
		SigGasConsumer:         ante.SigVerificationGasConsumer,
		MaxTxGasWanted:         0,
		TxFeeChecker:           ethante.NewDynamicFeeChecker(app.EvmKeeper),
		StakingKeeper:          app.StakingKeeper,
		DistributionKeeper:     app.DistrKeeper,
		ExtraDecorator:         poaante.NewPoaDecorator(),
		AuthzDisabledMsgTypes: []string{
			sdk.MsgTypeURL(&stakingtypes.MsgUndelegate{}),
			sdk.MsgTypeURL(&stakingtypes.MsgBeginRedelegate{}),
		},
	}
}

func (aa *AppAnteHandlerOptions) Validate() error {
	return (*ante.HandlerOptions)(aa).Validate()
}

func (aa *AppAnteHandlerOptions) Options() ante.HandlerOptions {
	return ante.HandlerOptions(*aa)
}

func (aa *AppAnteHandlerOptions) WithCodec(cdc codec.BinaryCodec) *AppAnteHandlerOptions {
	aa.Cdc = cdc
	return aa
}

func (aa *AppAnteHandlerOptions) WithSignModeHandler(signModeHandler authsigning.SignModeHandler) *AppAnteHandlerOptions {
	aa.SignModeHandler = signModeHandler
	return aa
}

func (aa *AppAnteHandlerOptions) WithMaxTxGasWanted(maxTxGasWanted uint64) *AppAnteHandlerOptions {
	aa.MaxTxGasWanted = maxTxGasWanted
	return aa
}
