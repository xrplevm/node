package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgRemoveValidator = "remove_validator"

var _ sdk.Msg = &MsgRemoveValidator{}

func NewMsgRemoveValidator(address string) *MsgRemoveValidator {
	return &MsgRemoveValidator{
		ValidatorAddress: address,
	}
}

func (msg *MsgRemoveValidator) Route() string {
	return RouterKey
}

func (msg *MsgRemoveValidator) Type() string {
	return TypeMsgRemoveValidator
}

func (msg *MsgRemoveValidator) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{addr}
}

func (msg *MsgRemoveValidator) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRemoveValidator) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return sdkerrors.Wrap(err, "authority")
	}
	if _, err := sdk.AccAddressFromBech32(msg.ValidatorAddress); err != nil {
		return sdkerrors.Wrap(err, "validator_address")
	}
	return nil
}
