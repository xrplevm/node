package types

import errorsmod "cosmossdk.io/errors"

var (
	ErrInvalidParams       = errorsmod.Register(ModuleName, 2, "invalid fastupgrade params")
	ErrIncompatibleWithGov = errorsmod.Register(ModuleName, 3, "fastupgrade threshold incompatible with gov params")
)
