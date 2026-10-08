package v12

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
	params evmtypes.Params
}

func (m *mockEvmKeeper) GetParams(sdk.Context) evmtypes.Params { return m.params }
func (m *mockEvmKeeper) SetParams(_ sdk.Context, params evmtypes.Params) error {
	m.params = params
	return nil
}

func TestCreateUpgradeHandlerDisablesGovPrecompile(t *testing.T) {
	ctx := sdk.NewContext(nil, cmtproto.Header{}, false, log.NewNopLogger())
	configurator := module.NewConfigurator(
		codec.NewProtoCodec(codectypes.NewInterfaceRegistry()),
		grpc.NewServer(),
		grpc.NewServer(),
	)
	keeper := &mockEvmKeeper{params: evmtypes.Params{
		ActiveStaticPrecompiles: []string{
			evmtypes.StakingPrecompileAddress,
			evmtypes.GovPrecompileAddress,
			evmtypes.BankPrecompileAddress,
		},
	}}

	_, err := CreateUpgradeHandler(module.NewManager(), configurator, keeper)(
		ctx,
		upgradetypes.Plan{Name: UpgradeName},
		module.VersionMap{},
	)

	require.NoError(t, err)
	require.Equal(t,
		[]string{evmtypes.StakingPrecompileAddress, evmtypes.BankPrecompileAddress},
		keeper.params.ActiveStaticPrecompiles,
	)
}
