package types

// DONTCOVER

import (
	sdkerrors "cosmossdk.io/errors"
)

// x/poa module sentinel errors
var (
	ErrAddressHasBankTokens      = sdkerrors.Register(ModuleName, 1, "address already has bank tokens")
	ErrAddressHasBondedTokens    = sdkerrors.Register(ModuleName, 2, "address already has bonded tokens")
	ErrAddressHasUnbondingTokens = sdkerrors.Register(ModuleName, 3, "address already has unbonding tokens")
	ErrAddressHasDelegatedTokens = sdkerrors.Register(ModuleName, 4, "address already has delegated tokens")
	ErrInvalidValidatorStatus    = sdkerrors.Register(ModuleName, 5, "invalid validator status")
	ErrAddressIsNotAValidator    = sdkerrors.Register(ModuleName, 6, "address is not a validator")
	ErrMaxValidatorsReached      = sdkerrors.Register(ModuleName, 7, "maximum number of validators reached")
)
