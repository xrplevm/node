package constants

import (
	testconstants "github.com/cosmos/evm/testutil/constants"
	evmtypes "github.com/cosmos/evm/x/vm/types"
)

const (
	DisplayDenom = "token"
	// BaseDenom defines to the default denomination used in EVM
	BaseDenom = "token"
	// BaseDenomUnit defines the unit of the base denomination
	BaseDenomUnit = 18
)

var (
	LocalnetChainID = testconstants.ChainID{
		ChainID:    "exrp_1449999-1",
		EVMChainID: 1449999,
	}

	ExampleChainCoinInfo = map[testconstants.ChainID]evmtypes.EvmCoinInfo{
		LocalnetChainID: {
			Denom:         BaseDenom,
			ExtendedDenom: BaseDenom,
			DisplayDenom:  DisplayDenom,
			Decimals:      BaseDenomUnit,
		},
	}
)
