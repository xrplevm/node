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
	"github.com/cosmos/cosmos-sdk/types/module"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
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

func TestCreateUpgradeHandler(t *testing.T) {
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

	_, err := CreateUpgradeHandler(module.NewManager(), configurator, keeper, keys)(
		ctx,
		upgradetypes.Plan{Name: UpgradeName},
		module.VersionMap{},
	)

	require.NoError(t, err)
	require.Equal(t,
		[]string{evmtypes.Bech32PrecompileAddress, evmtypes.BankPrecompileAddress},
		keeper.params.ActiveStaticPrecompiles,
	)
	for name, key := range keys {
		itr := ctx.KVStore(key).Iterator(nil, nil)
		require.False(t, itr.Valid(), "store %s is not empty", name)
		require.NoError(t, itr.Close())
	}
}
