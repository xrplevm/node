package v115

import (
	"slices"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/stretchr/testify/require"
)

// fakeEvmKeeper keeps the EVM params in memory.
type fakeEvmKeeper struct {
	params evmtypes.Params
}

// GetParams returns a copy, as the real keeper decodes the params from the
// store on each read.
func (k *fakeEvmKeeper) GetParams(sdk.Context) evmtypes.Params {
	params := k.params
	params.ActiveStaticPrecompiles = slices.Clone(params.ActiveStaticPrecompiles)
	return params
}

func (k *fakeEvmKeeper) SetParams(_ sdk.Context, params evmtypes.Params) error {
	k.params = params
	return nil
}

// mainnetParams returns the EVM params that mainnet (xrplevm_1440000-1) and
// testnet (xrplevm_1449000-1) hold before v11.5.0, as read from
// /cosmos/evm/vm/v1/params on 2026-11-24.
func mainnetParams() evmtypes.Params {
	return evmtypes.Params{
		EvmDenom:    "axrp",
		ExtraEIPs:   []int64{3855},
		EVMChannels: []string{},
		AccessControl: evmtypes.AccessControl{
			Create: evmtypes.AccessControlType{AccessType: evmtypes.AccessTypePermissionless, AccessControlList: []string{}},
			Call:   evmtypes.AccessControlType{AccessType: evmtypes.AccessTypePermissionless, AccessControlList: []string{}},
		},
		ActiveStaticPrecompiles: []string{
			"0x0000000000000000000000000000000000000100",
			"0x0000000000000000000000000000000000000400",
			"0x0000000000000000000000000000000000000804",
			"0x0000000000000000000000000000000000000805",
		},
	}
}

// TestRemoveGovAndP256Precompiles verifies that the mainnet params keep only
// the bech32 and bank precompiles, and that no other param changes.
func TestRemoveGovAndP256Precompiles(t *testing.T) {
	keeper := &fakeEvmKeeper{params: mainnetParams()}

	require.NoError(t, removeGovAndP256Precompiles(sdk.Context{}, keeper))

	expected := mainnetParams()
	expected.ActiveStaticPrecompiles = []string{
		evmtypes.Bech32PrecompileAddress,
		evmtypes.BankPrecompileAddress,
	}
	require.Equal(t, expected, keeper.params)
}
