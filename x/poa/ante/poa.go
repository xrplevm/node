package ante

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var ErrTxTypeNotAllowed = errors.New("tx type not allowed")

type PoaDecorator struct{}

func NewPoaDecorator() PoaDecorator {
	return PoaDecorator{}
}

var disallowedMsgs = map[string]struct{}{
	sdk.MsgTypeURL(&stakingtypes.MsgUndelegate{}):                {},
	sdk.MsgTypeURL(&stakingtypes.MsgBeginRedelegate{}):           {},
	sdk.MsgTypeURL(&stakingtypes.MsgDelegate{}):                  {},
	sdk.MsgTypeURL(&stakingtypes.MsgCancelUnbondingDelegation{}): {},
	sdk.MsgTypeURL(&stakingtypes.MsgCreateValidator{}):           {},
}

func (cbd PoaDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	// Allow genutil gentxs at genesis (block height 0) to bootstrap the initial validator set.
	if ctx.BlockHeight() == 0 {
		return next(ctx, tx, simulate)
	}
	for _, msg := range tx.GetMsgs() {
		if _, blocked := disallowedMsgs[sdk.MsgTypeURL(msg)]; blocked {
			return ctx, ErrTxTypeNotAllowed
		}
	}

	return next(ctx, tx, simulate)
}
