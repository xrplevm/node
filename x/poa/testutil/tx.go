package testutil

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	protov2 "google.golang.org/protobuf/proto"
)

type Tx interface {
	sdk.HasMsgs

	// GetMsgsV2 gets the transaction's messages as google.golang.org/protobuf/proto.Message's.
	GetMsgsV2() ([]protov2.Message, error)
}
