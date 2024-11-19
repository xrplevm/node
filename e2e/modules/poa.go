package modules

import (
	"github.com/xrplevm/node/v3/testutil/network"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

type PoaOps interface {
	ChangeValidator(action int, address sdk.AccAddress, pubKey cryptotypes.PubKey, validators []*network.Validator, waitStatus govtypesv1.ProposalStatus)
	BondTokens(validator *network.Validator, tokens sdk.Int)
	UnbondTokens(validator *network.Validator, tokens sdk.Int, wait bool) string
	Delegate(validator *network.Validator, tokens sdk.Int) string
	Redelegate(src *network.Validator, dst *network.Validator, tokens sdk.Int) string
}
