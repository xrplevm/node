package ante

import (
	"testing"
	"time"

	storetypes "cosmossdk.io/store/types"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/xrplevm/node/v10/x/poa/testutil"
	"github.com/xrplevm/node/v10/x/poa/types"
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
		blockHeight   int64
		msgs          []sdk.Msg
		expectedError error
	}{
		{
			name:          "should return error - rejects MsgUndelegate post-genesis",
			blockHeight:   1,
			msgs:          []sdk.Msg{&stakingtypes.MsgUndelegate{}},
			expectedError: ErrTxTypeNotAllowed,
		},
		{
			name:          "should return error - rejects MsgBeginRedelegate post-genesis",
			blockHeight:   1,
			msgs:          []sdk.Msg{&stakingtypes.MsgBeginRedelegate{}},
			expectedError: ErrTxTypeNotAllowed,
		},
		{
			name:          "should return error - rejects MsgDelegate post-genesis",
			blockHeight:   1,
			msgs:          []sdk.Msg{&stakingtypes.MsgDelegate{}},
			expectedError: ErrTxTypeNotAllowed,
		},
		{
			name:          "should return error - rejects MsgCancelUnbondingDelegation post-genesis",
			blockHeight:   1,
			msgs:          []sdk.Msg{&stakingtypes.MsgCancelUnbondingDelegation{}},
			expectedError: ErrTxTypeNotAllowed,
		},
		{
			name:          "should return error - rejects MsgCreateValidator post-genesis",
			blockHeight:   1,
			msgs:          []sdk.Msg{&stakingtypes.MsgCreateValidator{}},
			expectedError: ErrTxTypeNotAllowed,
		},
		{
			name:        "pass - allows MsgEditValidator post-genesis",
			blockHeight: 1,
			msgs:        []sdk.Msg{&stakingtypes.MsgEditValidator{}},
		},
		{
			name:        "pass - allows MsgCreateValidator at genesis",
			blockHeight: 0,
			msgs:        []sdk.Msg{&stakingtypes.MsgCreateValidator{}},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			pd, ctx := setupPoaDecorator(t)
			ctx = ctx.WithBlockHeight(tc.blockHeight)

			ctrl := gomock.NewController(t)
			txMock := testutil.NewMockTx(ctrl)
			txMock.EXPECT().GetMsgs().Return(tc.msgs).AnyTimes()

			mockNext := func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
				return ctx, nil
			}

			_, err := pd.AnteHandle(ctx, txMock, false, mockNext)
			if tc.expectedError != nil {
				require.ErrorIs(t, err, tc.expectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
