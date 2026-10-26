package app

import signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"

// EnabledSignModes lists the sign modes that the chain and the CLI accept.
// SIGN_MODE_DIRECT_AUX is not used on this chain, so it is not enabled.
// SIGN_MODE_DIRECT is first, so it is the default mode.
var EnabledSignModes = []signingtypes.SignMode{
	signingtypes.SignMode_SIGN_MODE_DIRECT,
	signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
}
