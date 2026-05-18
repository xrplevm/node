package v7

import (
	"testing"

	"cosmossdk.io/log"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestCreateUpgradeHandlerCompletes(t *testing.T) {
	ctx := sdk.NewContext(nil, cmtproto.Header{}, false, log.NewNopLogger())
	moduleManager := module.NewManager()
	configurator := module.NewConfigurator(
		codec.NewProtoCodec(codectypes.NewInterfaceRegistry()),
		grpc.NewServer(),
		grpc.NewServer(),
	)
	versions := module.VersionMap{"bank": 1}

	updatedVersions, err := CreateUpgradeHandler(moduleManager, configurator)(
		ctx,
		upgradetypes.Plan{Name: UpgradeName},
		versions,
	)

	require.NoError(t, err)
	require.Empty(t, updatedVersions)
}
