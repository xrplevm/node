package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgRemoveValidator{}

func NewMsgRemoveValidator(authority string, address string) *MsgRemoveValidator {
	return &MsgRemoveValidator{
		Authority:        authority,
		ValidatorAddress: address,
	}
}
