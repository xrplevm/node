package keeper_test

import (
	"testing"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

	v12 "github.com/xrplevm/node/v12/app/upgrades/v12"
	"github.com/xrplevm/node/v12/x/fastupgrade/types"
)

func TestModuleManager_FastUpgradeOrder(t *testing.T) {
	f := newFixture(t)
	mm := f.app.ModuleManager()
	index := func(order []string, name string) int {
		for i, n := range order {
			if n == name {
				return i
			}
		}
		t.Fatalf("%s not in order", name)
		return -1
	}
	require.Less(t, index(mm.OrderEndBlockers, types.ModuleName), index(mm.OrderEndBlockers, govtypes.ModuleName))
	require.Less(t, index(mm.OrderEndBlockers, govtypes.ModuleName), index(mm.OrderEndBlockers, stakingtypes.ModuleName))
	require.Less(t, index(mm.OrderInitGenesis, govtypes.ModuleName), index(mm.OrderInitGenesis, types.ModuleName))
}

func TestCreateUpgradeHandler_EnablesFastUpgrade(t *testing.T) {
	f := newFixtureWithoutBaseline(t)
	require.Equal(t, []string{types.StoreKey}, v12.StoreUpgrades().Added)

	params, err := f.k.Params.Get(f.ctx)
	require.NoError(t, err)
	require.Equal(t, types.DefaultParams(), params)

	t.Run("incompatible gov params leave the module disabled", func(t *testing.T) {
		ctx, _ := f.ctx.CacheContext()
		g := *f
		g.t = t
		g.ctx = ctx
		g.setGovParams(func(p *govv1.Params) { p.ExpeditedThreshold = incompatibleGovThreshold })
		g.applyUpgrade(v12.UpgradeName)
		params, err := g.k.Params.Get(g.ctx)
		require.NoError(t, err)
		require.Equal(t, types.DefaultParams(), params)
	})

	t.Run("a version map without the module runs its InitGenesis", func(t *testing.T) {
		ctx, _ := f.ctx.CacheContext()
		require.NoError(t, f.k.Params.Remove(ctx))
		vm, err := f.app.UpgradeKeeper.GetModuleVersionMap(ctx)
		require.NoError(t, err)
		delete(vm, types.ModuleName)

		handler := v12.CreateUpgradeHandler(f.app.ModuleManager(), f.app.Configurator(), f.k)
		vm, err = handler(ctx, upgradetypes.Plan{Name: v12.UpgradeName}, vm)
		require.NoError(t, err)
		require.Equal(t, uint64(1), vm[types.ModuleName])
		params, err := f.k.Params.Get(ctx)
		require.NoError(t, err)
		require.Equal(t, types.Params{Enabled: true, TimelockBlocks: 150, Threshold: "0.850000000000000000", PostponeBlocks: 10}, params)
	})

	f.applyUpgrade(v12.UpgradeName)
	params, err = f.k.Params.Get(f.ctx)
	require.NoError(t, err)
	require.Equal(t, types.Params{Enabled: true, TimelockBlocks: 150, Threshold: "0.850000000000000000", PostponeBlocks: 10}, params)

	vm, err := f.app.UpgradeKeeper.GetModuleVersionMap(f.ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(1), vm[types.ModuleName])

	name, height, err := f.app.UpgradeKeeper.GetLastCompletedUpgrade(f.ctx)
	require.NoError(t, err)
	require.Equal(t, v12.UpgradeName, name)
	require.Equal(t, startHeight, height)
}
