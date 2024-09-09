package ante

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/xrplevm/node/v3/x/poa/testutil"
)

func TestPoaDecorator_AnteHandle(t *testing.T) {
	pd, ctx := setupPoaDecorator(t)

	ctrl := gomock.NewController(t)
	txMock := testutil.NewMockTx(ctrl)
	txMock.EXPECT().GetMsgs().Return([]sdk.Msg{}).AnyTimes()

	mockNext := func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		return ctx, nil
	}

	_, err := pd.AnteHandle(ctx, txMock, false, mockNext)
	require.NoError(t, err)
}
