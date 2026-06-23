package v11

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	UpgradeName  = "v11.0.0"
	EVMCoinDenom = "axrp"

	// mainnet
	MainnetChainID    = "xrplevm_1440000-1"
	ElysChannelID     = "channel-1"
	WithdrawalAddress = "ethm1m2pp8zjwk3ystxyxvw5h3mrhhhnzcr2ltjntz9"

	// testnet
	TestnetChainID           = "xrplevm_1449000-1"
	TestnetElysChannelID     = "channel-17"
	TestnetWithdrawalAddress = "ethm16gt28px9q0fp48eatecp7j032lm5vaxs2t29pa"

	// devnet
	DevnetChainID           = "xrplevm_1449900-1"
	DevnetElysChannelID     = "channel-4"
	DevnetWithdrawalAddress = "ethm16gt28px9q0fp48eatecp7j032lm5vaxs2t29pa"
)

var (
	devnetAmount, _      = sdkmath.NewIntFromString("2000000000000000000")
	testnetAmount, _     = sdkmath.NewIntFromString("2000000000000000000")
	mainnetElysAmount, _ = sdkmath.NewIntFromString("6955539034646993768414")
)

// ElysRecovery holds, for a single network, the Elys transfer channel whose
// escrow holds the stranded XRP, the address that should receive it, and the
// coin (denom + amount of XRP in axrp base units) to unescrow.
type ElysRecovery struct {
	ChannelID         string
	WithdrawalAddress string
	Coin              sdk.Coin
}

// ElysRecoveryByNetwork maps each network's Cosmos chain ID to its Elys recovery
// parameters. The v11 handler selects the entry matching ctx.ChainID().
var ElysRecoveryByNetwork = map[string]ElysRecovery{
	MainnetChainID: {
		ChannelID:         ElysChannelID,
		WithdrawalAddress: WithdrawalAddress,
		Coin:              sdk.NewCoin(EVMCoinDenom, mainnetElysAmount),
	},
	TestnetChainID: {
		ChannelID:         TestnetElysChannelID,
		WithdrawalAddress: TestnetWithdrawalAddress,
		Coin:              sdk.NewCoin(EVMCoinDenom, testnetAmount),
	},
	DevnetChainID: {
		ChannelID:         DevnetElysChannelID,
		WithdrawalAddress: DevnetWithdrawalAddress,
		Coin:              sdk.NewCoin(EVMCoinDenom, devnetAmount),
	},
}
