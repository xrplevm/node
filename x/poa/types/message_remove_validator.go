package types

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const TypeMsgRemoveValidator = "remove_validator"

var _ sdk.Msg = &MsgRemoveValidator{}

func NewMsgRemoveValidator(authority string, address string) *MsgRemoveValidator {
	return &MsgRemoveValidator{
		Authority:        authority,
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
	addr, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg *MsgRemoveValidator) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRemoveValidator) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return errors.Wrap(err, "authority")
	}
	if _, err := sdk.AccAddressFromBech32(msg.ValidatorAddress); err != nil {
		return errors.Wrap(err, "validator_address")
	}
	return nil
}
