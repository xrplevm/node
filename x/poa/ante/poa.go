package ante

import (
	"errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type PoaDecorator struct{}

func NewPoaDecorator() PoaDecorator {
	return PoaDecorator{}
}

func (cbd PoaDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	// loop through all the messages and check if the message type is allowed
	for _, msg := range tx.GetMsgs() {
		if sdk.MsgTypeURL(msg) == "/cosmos.staking.v1beta1.MsgUndelegate" ||
			sdk.MsgTypeURL(msg) == "/cosmos.staking.v1beta1.MsgBeginRedelegate" {
			return ctx, errors.New("tx type not allowed")
		}
	}

	return next(ctx, tx, simulate)
}
