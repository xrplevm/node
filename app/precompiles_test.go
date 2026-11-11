package app_test

import (
	"slices"
	"testing"

	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/stretchr/testify/require"
	"github.com/xrplevm/node/v12/testutil/constants"
	exrpcommon "github.com/xrplevm/node/v12/testutil/integration/exrp/common"
)

// Only the Cosmos precompiles in use on mainnet and testnet are registered.
// The others call Cosmos keepers and msg servers directly, so activating one
// of their addresses must not make it callable.
func TestNew_RegisteredStaticPrecompiles(t *testing.T) {
	exrpApp := exrpcommon.CreateExrpApp(constants.LocalnetChainID.ChainID)

	// The static precompiles that app.New registers in the EVM keeper, next to
	// the Ethereum native ones.
	registered := []string{
		evmtypes.P256PrecompileAddress,
		evmtypes.Bech32PrecompileAddress,
		evmtypes.BankPrecompileAddress,
	}

	// AvailableStaticPrecompiles is the cosmos/evm list of every Cosmos static
	// precompile address. It holds the registered ones and the ones that panic.
	for _, address := range evmtypes.AvailableStaticPrecompiles {
		t.Run(address, func(t *testing.T) {
			params := evmtypes.Params{ActiveStaticPrecompiles: []string{address}}
			get := func() {
				_, found, err := exrpApp.EvmKeeper.GetStaticPrecompileInstance(&params, common.HexToAddress(address))
				require.NoError(t, err)
				require.True(t, found)
			}

			if slices.Contains(registered, address) {
				get()
				return
			}
			// The other addresses are not registered: staking, distribution, ICS20,
			// vesting, gov and slashing. The EVM keeper panics on an active address
			// with no registered precompile.
			require.Panics(t, get)
		})
	}

	// The Ethereum native precompiles are always active, so each one must be
	// registered. If one is missing, every call to it panics.
	for _, address := range vm.PrecompiledAddressesPrague {
		t.Run(address.Hex(), func(t *testing.T) {
			_, found, err := exrpApp.EvmKeeper.GetStaticPrecompileInstance(&evmtypes.Params{}, address)
			require.NoError(t, err)
			require.True(t, found)
		})
	}
}
