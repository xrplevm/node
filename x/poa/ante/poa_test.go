package ante

import (
	"errors"
	"testing"
	"time"

	storetypes "cosmossdk.io/store/types"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/xrplevm/node/v8/x/poa/testutil"
	"github.com/xrplevm/node/v8/x/poa/types"
)

func setupPoaDecorator(t *testing.T) (
	PoaDecorator,
	sdk.Context,
) {
	key := storetypes.NewKVStoreKey(types.StoreKey)
	tsKey := storetypes.NewTransientStoreKey("transient_test")
	testCtx := sdktestutil.DefaultContextWithDB(t, key, tsKey)
	ctx := testCtx.Ctx.WithBlockHeader(tmproto.Header{Time: time.Now()})

	return NewPoaDecorator(), ctx
}

func TestPoaDecorator_AnteHandle(t *testing.T) {
	tt := []struct {
		name          string
		msgs          []sdk.Msg
		expectedError error
	}{
		{
			name: "should return error - tx not allowed",
			msgs: []sdk.Msg{
				&stakingtypes.MsgUndelegate{},
				&stakingtypes.MsgBeginRedelegate{},
				&stakingtypes.MsgDelegate{},
				&stakingtypes.MsgCancelUnbondingDelegation{},
			},
			expectedError: errors.New("tx type not allowed"),
		},
		{
			name: "should not return error",
			msgs: []sdk.Msg{
				&stakingtypes.MsgEditValidator{},
			},
		},
	}

	for _, tc := range tt {
		pd, ctx := setupPoaDecorator(t)

		ctrl := gomock.NewController(t)
		txMock := testutil.NewMockTx(ctrl)
		txMock.EXPECT().GetMsgs().Return(tc.msgs).AnyTimes()

		mockNext := func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
			return ctx, nil
		}

		_, err := pd.AnteHandle(ctx, txMock, false, mockNext)
		if tc.expectedError != nil {
			require.Error(t, err)
			require.Equal(t, tc.expectedError, err)
		} else {
			require.NoError(t, err)
		}
	}
}
