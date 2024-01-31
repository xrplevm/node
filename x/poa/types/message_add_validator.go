package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

const TypeMsgAddValidator = "add_validator"

var (
	_ sdk.Msg                            = &MsgAddValidator{}
	_ codectypes.UnpackInterfacesMessage = (*MsgAddValidator)(nil)
)

func NewMsgAddValidator(authority string, address string, pubKey cryptotypes.PubKey, description stakingtypes.Description) (*MsgAddValidator, error) {
	var pkAny *codectypes.Any
	if pubKey != nil {
		var err error
		if pkAny, err = codectypes.NewAnyWithValue(pubKey); err != nil {
			return nil, err
		}
	}
	return &MsgAddValidator{
		Authority:        authority,
		ValidatorAddress: address,
		Pubkey:           pkAny,
		Description:      description,
	}, nil
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
	if msg.Pubkey == nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "nil pubkey")
	}
	if msg.Description == (stakingtypes.Description{}) {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "empty description")
	}
	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg *MsgAddValidator) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var pubKey cryptotypes.PubKey
	return unpacker.UnpackAny(msg.Pubkey, &pubKey)
}
