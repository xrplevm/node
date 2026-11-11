package keeper_test

import (
	"testing"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"

	"github.com/xrplevm/node/v12/x/fastupgrade/keeper"
	"github.com/xrplevm/node/v12/x/fastupgrade/types"
)

func TestUpdateParams(t *testing.T) {
	f := newFixture(t)
	srv := keeper.NewMsgServerImpl(f.k)
	valid := types.Params{Enabled: true, TimelockBlocks: 2, Threshold: "1", PostponeBlocks: 1}

	_, err := srv.UpdateParams(f.ctx, &types.MsgUpdateParams{Authority: f.other.String(), Params: valid})
	require.ErrorIs(t, err, govtypes.ErrInvalidSigner)

	_, err = srv.UpdateParams(f.ctx, &types.MsgUpdateParams{Authority: f.govAddr, Params: types.Params{TimelockBlocks: 1, Threshold: "0.85"}})
	require.ErrorIs(t, err, types.ErrInvalidParams)

	incompatible := types.Params{Enabled: true, TimelockBlocks: 150, Threshold: "0.6", PostponeBlocks: 10}
	_, err = srv.UpdateParams(f.ctx, &types.MsgUpdateParams{Authority: f.govAddr, Params: incompatible})
	require.ErrorIs(t, err, types.ErrIncompatibleWithGov)

	// Gov can always disable the module, also with an incompatible threshold.
	incompatible.Enabled = false
	_, err = srv.UpdateParams(f.ctx, &types.MsgUpdateParams{Authority: f.govAddr, Params: incompatible})
	require.NoError(t, err)

	_, err = srv.UpdateParams(f.ctx, &types.MsgUpdateParams{Authority: f.govAddr, Params: valid})
	require.NoError(t, err)
	res, err := keeper.NewQueryServerImpl(f.k).Params(f.ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, valid, res.Params)
}

func TestGenesisExportsOnlyParams(t *testing.T) {
	f := newFixture(t)
	id := f.proposeMinor()
	f.startTimer(id)

	gs, err := f.k.ExportGenesis(f.ctx)
	require.NoError(t, err)
	require.Equal(t, &types.GenesisState{Params: types.Params{Enabled: true, TimelockBlocks: 150, Threshold: "0.850000000000000000", PostponeBlocks: 10}}, gs)

	require.ErrorIs(t, f.k.InitGenesis(f.ctx, types.GenesisState{Params: types.Params{TimelockBlocks: 0, Threshold: "0.85"}}), types.ErrInvalidParams)

	// Gov compatibility is a runtime check, so an export taken after gov
	// params drifted still imports.
	drifted := types.GenesisState{Params: types.Params{Enabled: true, TimelockBlocks: 150, Threshold: "0.5", PostponeBlocks: 10}}
	require.NoError(t, f.k.InitGenesis(f.ctx, drifted))
}

func TestTimerQuery(t *testing.T) {
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
