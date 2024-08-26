package types

// DONTCOVER

import (
	sdkerrors "cosmossdk.io/errors"
)

// x/poa module sentinel errors
var (
	ErrAddressHasBankTokens      = sdkerrors.Register(ModuleName, 2, "address already has bank tokens")
	ErrAddressHasNoTokens        = sdkerrors.Register(ModuleName, 3, "address has not tokens")
	ErrAddressHasBondedTokens    = sdkerrors.Register(ModuleName, 4, "address already has bonded tokens")
	ErrAddressHasUnbondingTokens = sdkerrors.Register(ModuleName, 5, "address already has unbonding tokens")
	ErrAddressHasUnbondedTokens  = sdkerrors.Register(ModuleName, 6, "address already has unbonded tokens")
	ErrAddressHasDelegatedTokens = sdkerrors.Register(ModuleName, 7, "address already has delegated tokens")
	ErrInvalidValidatorStatus    = sdkerrors.Register(ModuleName, 8, "invalid validator status")
)
