package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg              = &MsgRemoveValidator{}
	_ sdk.HasValidateBasic = &MsgRemoveValidator{}
)

func NewMsgRemoveValidator(authority string, address string) *MsgRemoveValidator {
	return &MsgRemoveValidator{
		Authority:        authority,
		ValidatorAddress: address,
	}
}

// ValidateBasic performs stateless validation of the message fields.
func (msg *MsgRemoveValidator) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid authority address: %s", err)
	}
	if _, err := sdk.AccAddressFromBech32(msg.ValidatorAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}
	return nil
}
