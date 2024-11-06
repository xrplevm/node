package testutil

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

type PubKey interface {
	cryptotypes.PubKey
}
