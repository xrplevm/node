package ante

import (
	"testing"
	"time"

	storetypes "cosmossdk.io/store/types"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/xrplevm/node/v4/x/poa/testutil"
	"github.com/xrplevm/node/v4/x/poa/types"
)

func setupPoaDecorator(t *testing.T) (
	*PoaDecorator,
	sdk.Context,
) {
	key := storetypes.NewKVStoreKey(types.StoreKey)
	tsKey := storetypes.NewTransientStoreKey("transient_test")
	testCtx := sdktestutil.DefaultContextWithDB(t, key, tsKey)
	ctx := testCtx.Ctx.WithBlockHeader(tmproto.Header{Time: time.Now()})

	return &PoaDecorator{}, ctx
}

func TestPoaDecorator_AnteHandle(t *testing.T) {
	pd, ctx := setupPoaDecorator(t)

	ctrl := gomock.NewController(t)
	txMock := testutil.NewMockTx(ctrl)
	txMock.EXPECT().GetMsgs().Return([]sdk.Msg{}).AnyTimes()

	mockNext := func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
		return ctx, nil
	}

	_, err := pd.AnteHandle(ctx, txMock, false, mockNext)
	require.NoError(t, err)
}
