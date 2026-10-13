package app_test

import (
	"encoding/json"
	"testing"

	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/stretchr/testify/require"
	"github.com/xrplevm/node/v12/app"
	"github.com/xrplevm/node/v12/testutil/constants"
	exrpcommon "github.com/xrplevm/node/v12/testutil/integration/exrp/common"
)

// Mainnet and testnet have the P256, Bech32 and Bank precompiles active.
func TestGenesis_ActiveStaticPrecompiles(t *testing.T) {
	exrpApp := exrpcommon.CreateExrpApp(constants.LocalnetChainID.ChainID)

	tt := []struct {
		name    string
		genesis map[string]json.RawMessage
	}{
		{
			name:    "App.DefaultGenesis",
			genesis: exrpApp.DefaultGenesis(),
		},
		{
			name:    "NewDefaultGenesisState",
			genesis: app.NewDefaultGenesisState(exrpApp),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var evmGenState evmtypes.GenesisState
			exrpApp.AppCodec().MustUnmarshalJSON(tc.genesis[evmtypes.ModuleName], &evmGenState)

			require.Equal(t, []string{
				evmtypes.P256PrecompileAddress,
				evmtypes.Bech32PrecompileAddress,
				evmtypes.BankPrecompileAddress,
			}, evmGenState.Params.ActiveStaticPrecompiles)
		})
	}
}
