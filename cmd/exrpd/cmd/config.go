package cmd

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethermint "github.com/evmos/evmos/v20/types"
)

const (
	// DisplayDenom defines the denomination displayed to users in client applications.
	DisplayDenom = "xrp"
	// BaseDenom defines to the default denomination used in EVM
	BaseDenom = "axrp"
)

func registerDenoms() {
	if err := sdk.RegisterDenom(DisplayDenom, math.LegacyOneDec()); err != nil {
		panic(err)
	}

	if err := sdk.RegisterDenom(BaseDenom, math.LegacyNewDecWithPrec(1, ethermint.BaseDenomUnit)); err != nil {
		panic(err)
	}
}
