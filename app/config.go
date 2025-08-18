package app

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	evmtypes "github.com/cosmos/evm/x/vm/types"
)

type EVMOptionsFn func(uint64) error

func NoOpEVMOptions(_ uint64) error {
	return nil
}

var (
	sealed             = false
	DefaultEvmCoinInfo = evmtypes.EvmCoinInfo{
		Denom:         BaseDenom,
		DisplayDenom:  DisplayDenom,
		ExtendedDenom: BaseDenom,
		Decimals:      evmtypes.EighteenDecimals,
	}
	DefaultLocalnetChainID   uint64 = 1440002
	DefaultSimulationChainID uint64 = 777
)

// ChainsCoinInfo maps EVM chain IDs to coin configuration
// IMPORTANT: Uses uint64 EVM chain IDs as keys, not Cosmos chain ID strings
var ChainsCoinInfo = map[uint64]evmtypes.EvmCoinInfo{
	DefaultLocalnetChainID:   DefaultEvmCoinInfo,
	DefaultSimulationChainID: DefaultEvmCoinInfo,
}

// EVMAppOptions sets up global configuration
func EVMAppOptions(chainID uint64) error {
	if sealed {
		return nil
	}

	// IMPORTANT: Lookup uses numeric EVMChainID, not Cosmos chainID string
	coinInfo, found := ChainsCoinInfo[chainID]
	if !found {
		return fmt.Errorf("unknown EVM chain id: %d", chainID)
	}

	//// Set denom info for the chain
	//if err := setBaseDenom(coinInfo); err != nil {
	//	return err
	//}

	ethCfg := evmtypes.DefaultChainConfig(chainID)

	err := evmtypes.NewEVMConfigurator().
		WithChainConfig(ethCfg).
		WithEVMCoinInfo(coinInfo).
		Configure()
	if err != nil {
		return err
	}

	sealed = true
	return nil
}

// setBaseDenom registers display and base denoms
func setBaseDenom(ci evmtypes.EvmCoinInfo) error {
	if err := sdk.RegisterDenom(ci.DisplayDenom, math.LegacyOneDec()); err != nil {
		return err
	}
	return sdk.RegisterDenom(ci.Denom, math.LegacyNewDecWithPrec(1, int64(ci.Decimals)))
}
