package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	_ sdk.Msg                            = &MsgAddValidator{}
	_ sdk.HasValidateBasic               = &MsgAddValidator{}
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

// ValidateBasic performs stateless validation of the message fields.
func (msg *MsgAddValidator) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid authority address: %s", err)
	}
	if _, err := sdk.AccAddressFromBech32(msg.ValidatorAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}
	if msg.Pubkey == nil {
		return sdkerrors.ErrInvalidPubKey.Wrap("validator pubkey is required")
	}
	if _, ok := msg.Pubkey.GetCachedValue().(cryptotypes.PubKey); !ok {
		return sdkerrors.ErrInvalidPubKey.Wrapf("expecting cryptotypes.PubKey, got %T", msg.Pubkey.GetCachedValue())
	}
	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg *MsgAddValidator) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var pubKey cryptotypes.PubKey
	return unpacker.UnpackAny(msg.Pubkey, &pubKey)
}
