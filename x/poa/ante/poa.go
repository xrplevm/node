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

// MsgCreateValidator stays routed for genesis gentxs and for x/poa, which
// creates validators through the msg router. Users must not send it. The other
// staking delegation msgs are not routed at all (app/msg_filter.go), so baseapp
// rejects them before the ante handler runs.
func (cbd PoaDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	// Allow genutil gentxs at genesis (block height 0) to bootstrap the initial validator set.
	if ctx.BlockHeight() == 0 {
		return next(ctx, tx, simulate)
	}
	for _, msg := range tx.GetMsgs() {
		if _, blocked := msg.(*stakingtypes.MsgCreateValidator); blocked {
			return ctx, ErrTxTypeNotAllowed
		}
	}

	return next(ctx, tx, simulate)
}
