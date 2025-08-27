package app

import (
	"fmt"

	evmtypes "github.com/cosmos/evm/x/vm/types"
)

type EVMOptionsFn func(uint64) error

func NoOpEVMOptions(_ uint64) error {
	return nil
}

var (
	// sealed specifies if the EVMConfigurator has been sealed or not.
	sealed = false

	DefaultEvmCoinInfo = evmtypes.EvmCoinInfo{
		Denom:         BaseDenom,
		DisplayDenom:  DisplayDenom,
		ExtendedDenom: BaseDenom,
		Decimals:      evmtypes.EighteenDecimals,
	}

	XrpEvmCoinInfo = evmtypes.EvmCoinInfo{
		Denom:         XrpDenom,
		DisplayDenom:  XrpDisplayDenom,
		ExtendedDenom: XrpDenom,
		Decimals:      evmtypes.EighteenDecimals,
	}

	LocalnetEVMChainID uint64 = 1440002
	MainnetEVMChainID  uint64 = 1440000
	TestnetEVMChainID  uint64 = 1449000
	DevnetEVMChainID   uint64 = 1440002

	SimulationEVMChainID uint64 = 777
)

// ChainsCoinInfo maps EVM chain IDs to coin configuration
// IMPORTANT: Uses uint64 EVM chain IDs as keys, not Cosmos chain ID strings
var ChainsCoinInfo = map[uint64]evmtypes.EvmCoinInfo{
	MainnetEVMChainID:    XrpEvmCoinInfo,
	TestnetEVMChainID:    XrpEvmCoinInfo,
	DevnetEVMChainID:     XrpEvmCoinInfo,
	LocalnetEVMChainID:   DefaultEvmCoinInfo,
	SimulationEVMChainID: DefaultEvmCoinInfo,
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
	// if err := setBaseDenom(coinInfo); err != nil {
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
