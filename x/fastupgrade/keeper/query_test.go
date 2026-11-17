package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/xrplevm/node/v12/x/fastupgrade/keeper"
	"github.com/xrplevm/node/v12/x/fastupgrade/types"
)

func TestQueryServer_Timer(t *testing.T) {
	f := newFixture(t)
	q := keeper.NewQueryServerImpl(f.k)

	res, err := q.Timer(f.ctx, &types.QueryTimerRequest{})
	require.NoError(t, err)
	require.Nil(t, res.Timer)

	timer := f.startTimer(f.proposeMinor())
	res, err = q.Timer(f.ctx, &types.QueryTimerRequest{})
	require.NoError(t, err)
	require.Equal(t, &timer, res.Timer)
}
