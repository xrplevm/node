package app

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	evmante "github.com/cosmos/evm/evmd/ante"
)

type HandlerOptions struct {
	evmante.HandlerOptions
	StakingKeeper          StakingKeeper
	DistributionKeeper     DistributionKeeper
	ExtraDecorator         sdk.AnteDecorator
	AuthzDisabledMsgTypes  []string
}

// Validate checks if the keepers are defined
func (options *HandlerOptions) Validate() error {
	if options.StakingKeeper == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "staking keeper is required for AnteHandler")
	}
	if options.DistributionKeeper == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "distribution keeper is required for AnteHandler")
	}
	return options.HandlerOptions.Validate()
}

func NewAnteHandler(options HandlerOptions) sdk.AnteHandler {
	return evmante.NewAnteHandler(options.HandlerOptions)
}