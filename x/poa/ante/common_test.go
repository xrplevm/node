package ante

import (
	"testing"
	"time"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/xrplevm/node/v3/x/poa/types"
)

func setupPoaDecorator(t *testing.T) (
	*PoaDecorator,
	sdk.Context,
) {
	key := sdk.NewKVStoreKey(types.StoreKey)
	tsKey := sdk.NewTransientStoreKey("transient_test")
	testCtx := sdktestutil.DefaultContextWithDB(t, key, tsKey)
	ctx := testCtx.Ctx.WithBlockHeader(tmproto.Header{Time: time.Now()})

	return &PoaDecorator{}, ctx
}
