package types

import errorsmod "cosmossdk.io/errors"

// Codes start at 2, as in the SDK's own modules. The SDK reserves code 1
// only in its "undefined" codespace, for internal errors.
var (
	ErrInvalidParams       = errorsmod.Register(ModuleName, 2, "invalid fastupgrade params")
	ErrIncompatibleWithGov = errorsmod.Register(ModuleName, 3, "fastupgrade threshold incompatible with gov params")
)
