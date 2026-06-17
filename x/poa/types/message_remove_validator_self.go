package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgRemoveValidatorSelf{}

func NewMsgRemoveValidatorSelf(address string) *MsgRemoveValidatorSelf {
	return &MsgRemoveValidatorSelf{
		ValidatorAddress: address,
	}
}
