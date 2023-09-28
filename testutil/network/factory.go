package network

import (
	simappparams "cosmossdk.io/simapp/params"
	"github.com/Peersyst/exrp/app"
	tmdb "github.com/cometbft/cometbft-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	pruningtypes "github.com/cosmos/cosmos-sdk/store/pruning/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

// NewAppConstructor returns a new simapp AppConstructor
func NewAppConstructor(encodingCfg simappparams.EncodingConfig) AppConstructor {
	return func(val Validator) servertypes.Application {
		return app.New(
			val.Ctx.Logger, tmdb.NewMemDB(), nil, true, make(map[int64]bool), val.Ctx.Config.RootDir, 0,
			encodingCfg,
			simtestutil.EmptyAppOptions{},
			baseapp.SetPruning(pruningtypes.NewPruningOptionsFromString(val.AppConfig.Pruning)),
			baseapp.SetMinGasPrices(val.AppConfig.MinGasPrices),
		)
	}
}

// NewTestNetwork creates instance with fully configured cosmos network.
// Accepts optional config, that will be used in place of the DefaultConfig() if provided.
func NewTestNetwork(t *testing.T, numValidators int, numBondedValidators int) *Network {
	cfg := DefaultConfig(numValidators, numBondedValidators)
	net, err := New(t, t.TempDir(), cfg)
	require.NoError(t, err)

	_, err = net.WaitForHeightWithTimeout(3, time.Minute)
	require.NoError(t, err)

	// t.Cleanup(net.Cleanup)
	return net
}
