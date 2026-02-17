package exrpcommon

import (
	"github.com/cosmos/evm/testutil/integration/evm/network"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	testconstants "github.com/xrplevm/node/v9/testutil/constants"
)

var defaultChain = testconstants.LocalnetChainID

type ChainCoins struct {
	// decimals of the base denom? Maybe not..
	baseCoin *network.CoinInfo
	evmCoin  *network.CoinInfo
}

func DefaultChainCoins() ChainCoins {
	baseCoinInfo := testconstants.ExampleChainCoinInfo[defaultChain]

	baseCoin := getCoinInfo(baseCoinInfo)
	evmCoin := getCoinInfo(baseCoinInfo)

	return ChainCoins{
		baseCoin: &baseCoin,
		evmCoin:  &evmCoin,
	}
}

func getCoinInfo(coinInfo evmtypes.EvmCoinInfo) network.CoinInfo {
	return network.CoinInfo{
		Denom:    coinInfo.Denom,
		Decimals: evmtypes.Decimals(coinInfo.Decimals),
	}
}

func (cc ChainCoins) BaseCoin() network.CoinInfo {
	return *cc.baseCoin
}

func (cc ChainCoins) EVMCoin() network.CoinInfo {
	return *cc.evmCoin
}

func (cc ChainCoins) BaseDenom() string {
	return cc.baseCoin.Denom
}

func (cc ChainCoins) EVMDenom() string {
	return cc.evmCoin.Denom
}

func (cc ChainCoins) BaseDecimals() evmtypes.Decimals {
	return cc.baseCoin.Decimals
}

func (cc ChainCoins) EVMDecimals() evmtypes.Decimals {
	return cc.evmCoin.Decimals
}

func (cc ChainCoins) IsBaseEqualToEVM() bool {
	return cc.BaseDenom() == cc.EVMDenom()
}

// DenomDecimalsMap returns a map of unique Denom -> Decimals for the chain
// coins.
func (cc ChainCoins) DenomDecimalsMap() map[string]evmtypes.Decimals {
	chainDenomDecimals := map[string]evmtypes.Decimals{
		cc.BaseDenom(): cc.BaseDecimals(),
	}

	// Insert the evm denom if base and evm denom are different.
	if !cc.IsBaseEqualToEVM() {
		chainDenomDecimals[cc.EVMDenom()] = cc.EVMDecimals()
	}
	return chainDenomDecimals
}
