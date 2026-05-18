package v10

import (
	"testing"

	"cosmossdk.io/log"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

var _ EvmKeeper = (*mockEvmKeeper)(nil)

type mockEvmKeeper struct {
	initErr    error
	initCalled bool
}

func (m *mockEvmKeeper) GetParams(sdk.Context) evmtypes.Params        { return evmtypes.Params{} }
func (m *mockEvmKeeper) SetParams(sdk.Context, evmtypes.Params) error { return nil }
func (m *mockEvmKeeper) SetCodeHash(sdk.Context, []byte, []byte)      {}
func (m *mockEvmKeeper) InitEvmCoinInfo(sdk.Context) error {
	m.initCalled = true
	return m.initErr
}

func TestCreateUpgradeHandlerInitializesEvmCoinInfo(t *testing.T) {
	ctx := sdk.NewContext(nil, cmtproto.Header{}, false, log.NewNopLogger())
	moduleManager := module.NewManager()
	configurator := module.NewConfigurator(
		codec.NewProtoCodec(codectypes.NewInterfaceRegistry()),
		grpc.NewServer(),
		grpc.NewServer(),
	)
	keeper := &mockEvmKeeper{}
	versions := module.VersionMap{"bank": 1}

	updatedVersions, err := CreateUpgradeHandler(moduleManager, configurator, keeper)(
		ctx,
		upgradetypes.Plan{Name: UpgradeName},
		versions,
	)

	require.NoError(t, err)
	require.True(t, keeper.initCalled)
	require.Empty(t, updatedVersions)
}
