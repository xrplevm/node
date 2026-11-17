package v12

import (
	"testing"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/cosmos-sdk/types/module"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	fastupgradetypes "github.com/xrplevm/node/v12/x/fastupgrade/types"
)

var _ EvmKeeper = (*mockEvmKeeper)(nil)

type mockEvmKeeper struct {
	params evmtypes.Params
}

func (m *mockEvmKeeper) GetParams(sdk.Context) evmtypes.Params { return m.params }
func (m *mockEvmKeeper) SetParams(_ sdk.Context, params evmtypes.Params) error {
	m.params = params
	return nil
}

var _ FastUpgradeKeeper = (*mockFastUpgradeKeeper)(nil)

// mockFastUpgradeKeeper stores the params like the real keeper. With
// incompatibleGov it refuses enabled params and stores nothing, as the real
// keeper does when gov's params do not fit the threshold.
type mockFastUpgradeKeeper struct {
	incompatibleGov bool
	params          fastupgradetypes.Params
}

func (m *mockFastUpgradeKeeper) SetParams(_ sdk.Context, params fastupgradetypes.Params) error {
	if m.incompatibleGov && params.Enabled {
		return fastupgradetypes.ErrIncompatibleWithGov
	}
	m.params = params
	return nil
}

func TestCreateUpgradeHandler(t *testing.T) {
	tt := []struct {
		name            string
		incompatibleGov bool
		expectedEnabled bool
	}{
		{
			name:            "should enable fastupgrade - compatible gov params",
			expectedEnabled: true,
		},
		{
			name:            "should keep fastupgrade disabled with the proposer - incompatible gov params",
			incompatibleGov: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			keys := storetypes.NewKVStoreKeys(RetiredStoreKeys...)
			ctx := sdktestutil.DefaultContextWithKeys(keys, nil, nil).WithLogger(log.NewNopLogger())
			for _, key := range keys {
				ctx.KVStore(key).Set([]byte("a"), []byte{1})
				ctx.KVStore(key).Set([]byte("b"), []byte{2})
			}
			configurator := module.NewConfigurator(
				codec.NewProtoCodec(codectypes.NewInterfaceRegistry()),
				grpc.NewServer(),
				grpc.NewServer(),
			)
			keeper := &mockEvmKeeper{params: evmtypes.Params{
				ActiveStaticPrecompiles: []string{
					evmtypes.Bech32PrecompileAddress,
					evmtypes.GovPrecompileAddress,
					evmtypes.BankPrecompileAddress,
				},
			}}
			fastUpgradeKeeper := &mockFastUpgradeKeeper{incompatibleGov: tc.incompatibleGov}

			_, err := CreateUpgradeHandler(module.NewManager(), configurator, keeper, keys, fastUpgradeKeeper)(
				ctx,
				upgradetypes.Plan{Name: UpgradeName},
				module.VersionMap{},
			)

			require.NoError(t, err)
			require.Equal(t,
				[]string{evmtypes.Bech32PrecompileAddress, evmtypes.BankPrecompileAddress},
				keeper.params.ActiveStaticPrecompiles,
			)
			require.Equal(t,
				fastupgradetypes.Params{
					Enabled: tc.expectedEnabled, TimelockBlocks: 150, Threshold: "0.850000000000000000", PostponeBlocks: 10,
					Proposers: []string{FastUpgradeProposer},
				},
				fastUpgradeKeeper.params,
			)
			// The real SetParams rejects an invalid address, and the upgrade then
			// fails. The mock does not validate, so the constant is checked here.
			prefix, _, err := bech32.DecodeAndConvert(FastUpgradeProposer)
			require.NoError(t, err)
			require.Equal(t, "ethm", prefix)
			for name, key := range keys {
				itr := ctx.KVStore(key).Iterator(nil, nil)
				require.False(t, itr.Valid(), "store %s is not empty", name)
				require.NoError(t, itr.Close())
			}
		})
	}
}
