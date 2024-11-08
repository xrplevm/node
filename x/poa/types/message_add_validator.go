package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	_ sdk.Msg = &MsgAddValidator{}
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
