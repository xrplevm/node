package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgAddValidator = "add_validator"

var _ sdk.Msg = &MsgAddValidator{}

func NewMsgAddValidator(address string) *MsgAddValidator {
	return &MsgAddValidator{
		ValidatorAddress: address,
	}
}

func (msg *MsgAddValidator) Route() string {
	return RouterKey
}

func (msg *MsgAddValidator) Type() string {
	return TypeMsgAddValidator
}

func (msg *MsgAddValidator) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg *MsgAddValidator) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgAddValidator) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return sdkerrors.Wrap(err, "authority")
	}
	if _, err := sdk.AccAddressFromBech32(msg.ValidatorAddress); err != nil {
		return sdkerrors.Wrap(err, "validator_address")
	}
	return nil
}
