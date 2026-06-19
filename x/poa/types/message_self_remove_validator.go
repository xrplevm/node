package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgSelfRemoveValidator{}

func NewMsgSelfRemoveValidator(address string) *MsgSelfRemoveValidator {
	return &MsgSelfRemoveValidator{
		Address: address,
	}
}
