package testutil

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Tx interface {
	GetMsgs() []sdk.Msg
	ValidateBasic() error
}
