package network

import (
	"github.com/Peersyst/exrp/app"
	appparams "github.com/Peersyst/exrp/app/params"
	"github.com/cosmos/cosmos-sdk/baseapp"
	pruningtypes "github.com/cosmos/cosmos-sdk/pruning/types"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/stretchr/testify/require"
	tmdb "github.com/tendermint/tm-db"
	"testing"
	"time"
)

// NewAppConstructor returns a new simapp AppConstructor
func NewAppConstructor(encodingCfg appparams.EncodingConfig) AppConstructor {
	return func(val Validator) servertypes.Application {
		return app.New(
			val.Ctx.Logger, tmdb.NewMemDB(), nil, true, make(map[int64]bool), val.Ctx.Config.RootDir, 0,
			encodingCfg,
			simapp.EmptyAppOptions{},
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
