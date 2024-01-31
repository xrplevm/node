package e2e

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

func GenPubKey() cryptotypes.PubKey {
	return ed25519.GenPrivKey().PubKey()
}
